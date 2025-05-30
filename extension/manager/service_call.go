package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/sony/gobreaker"
)

// DefaultCallOptions returns default call options
func DefaultCallOptions() *types.CallOptions {
	return &types.CallOptions{
		Strategy: types.LocalFirst,
		Timeout:  30 * time.Second,
	}
}

// CallService provides unified service calling interface
func (m *Manager) CallService(ctx context.Context, serviceName, methodName string, req any) (*types.CallResult, error) {
	return m.CallServiceWithOptions(ctx, serviceName, methodName, req, DefaultCallOptions())
}

// CallServiceWithOptions calls service
func (m *Manager) CallServiceWithOptions(ctx context.Context, serviceName, methodName string, req any, opts *types.CallOptions) (*types.CallResult, error) {
	if opts == nil {
		opts = DefaultCallOptions()
	}

	// Apply timeout
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	start := time.Now()
	var result *types.CallResult
	var err error

	switch opts.Strategy {
	case types.LocalFirst:
		result, err = m.callLocalFirst(ctx, serviceName, methodName, req, start)
	case types.RemoteFirst:
		result, err = m.callRemoteFirst(ctx, serviceName, methodName, req, start)
	case types.LocalOnly:
		result, err = m.callLocalOnly(ctx, serviceName, methodName, req, start)
	case types.RemoteOnly:
		result, err = m.callRemoteOnly(ctx, serviceName, methodName, req, start)
	default:
		result, err = m.callLocalFirst(ctx, serviceName, methodName, req, start)
	}

	// Track service call metrics
	duration := time.Since(start)
	m.trackServiceCall(serviceName, methodName, duration, err)

	return result, err
}

// callLocalFirst attempts local first, fallback to gRPC
func (m *Manager) callLocalFirst(ctx context.Context, serviceName, methodName string, req any, start time.Time) (*types.CallResult, error) {
	// Try local call
	if resp, err := m.callLocal(ctx, serviceName, methodName, req); err == nil {
		return &types.CallResult{
			Response: resp,
			IsLocal:  true,
			Duration: time.Since(start),
		}, nil
	}

	// Fallback to remote
	if resp, err := m.callRemote(ctx, serviceName, methodName, req); err == nil {
		return &types.CallResult{
			Response: resp,
			IsRemote: true,
			Duration: time.Since(start),
		}, nil
	}

	return &types.CallResult{
		Error:    fmt.Errorf("service %s.%s unavailable", serviceName, methodName),
		Duration: time.Since(start),
	}, fmt.Errorf("service %s.%s unavailable", serviceName, methodName)
}

// callRemoteFirst attempts gRPC first, fallback to local
func (m *Manager) callRemoteFirst(ctx context.Context, serviceName, methodName string, req any, start time.Time) (*types.CallResult, error) {
	// Try remote call
	if resp, err := m.callRemote(ctx, serviceName, methodName, req); err == nil {
		return &types.CallResult{
			Response: resp,
			IsRemote: true,
			Duration: time.Since(start),
		}, nil
	}

	// Fallback to local
	if resp, err := m.callLocal(ctx, serviceName, methodName, req); err == nil {
		return &types.CallResult{
			Response: resp,
			IsLocal:  true,
			Duration: time.Since(start),
		}, nil
	}

	return &types.CallResult{
		Error:    fmt.Errorf("service %s.%s unavailable", serviceName, methodName),
		Duration: time.Since(start),
	}, fmt.Errorf("service %s.%s unavailable", serviceName, methodName)
}

// callLocalOnly calls local service only
func (m *Manager) callLocalOnly(ctx context.Context, serviceName, methodName string, req any, start time.Time) (*types.CallResult, error) {
	resp, err := m.callLocal(ctx, serviceName, methodName, req)
	return &types.CallResult{
		Response: resp,
		Error:    err,
		IsLocal:  err == nil,
		Duration: time.Since(start),
	}, err
}

// callRemoteOnly calls gRPC service only
func (m *Manager) callRemoteOnly(ctx context.Context, serviceName, methodName string, req any, start time.Time) (*types.CallResult, error) {
	resp, err := m.callRemote(ctx, serviceName, methodName, req)
	return &types.CallResult{
		Response: resp,
		Error:    err,
		IsRemote: err == nil,
		Duration: time.Since(start),
	}, err
}

// callLocal handles local service calls
func (m *Manager) callLocal(ctx context.Context, serviceName, methodName string, req any) (any, error) {
	// Try cross service first
	if service, err := m.GetCrossService(serviceName, methodName); err == nil {
		return m.invokeServiceMethod(ctx, service, methodName, req)
	}

	// Try direct service access
	extensionService, err := m.GetServiceByName(serviceName)
	if err != nil {
		return nil, fmt.Errorf("local service %s not found: %v", serviceName, err)
	}

	return m.invokeServiceMethod(ctx, extensionService, methodName, req)
}

// callRemote handles remote gRPC service calls
func (m *Manager) callRemote(ctx context.Context, serviceName, methodName string, req any) (any, error) {
	if m.grpcRegistry == nil {
		return nil, fmt.Errorf("gRPC not enabled")
	}

	_, err := m.grpcRegistry.GetConnection(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get gRPC connection for %s: %v", serviceName, err)
	}

	// Convert request to JSON for generic transmission
	_, err = json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	logger.Debugf(ctx, "gRPC call: %s.%s", serviceName, methodName)

	// In a real implementation, this would make actual gRPC call
	return map[string]any{
		"service": serviceName,
		"method":  methodName,
		"request": req,
		"source":  "grpc",
	}, nil
}

// invokeServiceMethod uses reflection to call service methods
func (m *Manager) invokeServiceMethod(ctx context.Context, service any, methodName string, req any) (any, error) {
	serviceValue := reflect.ValueOf(service)

	// Handle pointers
	if serviceValue.Kind() == reflect.Ptr {
		if serviceValue.IsNil() {
			return nil, fmt.Errorf("service is nil")
		}
		serviceValue = serviceValue.Elem()
	}

	// Try to find the method
	methodValue := serviceValue.MethodByName(methodName)
	if !methodValue.IsValid() {
		// Try to find in the underlying struct
		if serviceValue.Kind() == reflect.Struct {
			methodValue = reflect.ValueOf(service).MethodByName(methodName)
		}

		if !methodValue.IsValid() {
			return nil, fmt.Errorf("method %s not found in service", methodName)
		}
	}

	// Prepare method arguments
	methodType := methodValue.Type()
	numIn := methodType.NumIn()

	args := make([]reflect.Value, numIn)
	argIndex := 0

	// Add context if method expects it
	if numIn > 0 && methodType.In(0).String() == "context.Context" {
		args[argIndex] = reflect.ValueOf(ctx)
		argIndex++
	}

	// Add request if method expects it
	if argIndex < numIn {
		args[argIndex] = reflect.ValueOf(req)
		argIndex++
	}

	// Fill remaining args with zero values if needed
	for i := argIndex; i < numIn; i++ {
		args[i] = reflect.Zero(methodType.In(i))
	}

	// Call the method
	results := methodValue.Call(args)

	// Process results
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		// Single return value - could be result or error
		result := results[0]
		if result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !result.IsNil() {
				return nil, result.Interface().(error)
			}
			return nil, nil
		}
		return result.Interface(), nil
	case 2:
		// Standard (result, error) pattern
		result := results[0]
		errResult := results[1]

		var err error
		if !errResult.IsNil() {
			err = errResult.Interface().(error)
		}

		if result.IsValid() && !result.IsNil() {
			return result.Interface(), err
		}

		return nil, err
	default:
		// Multiple return values - return as slice
		resultSlice := make([]any, len(results))
		for i, result := range results {
			if result.IsValid() && result.CanInterface() {
				resultSlice[i] = result.Interface()
			}
		}
		return resultSlice, nil
	}
}

// GetCrossService gets service by extension name and service path
func (m *Manager) GetCrossService(extensionName, servicePath string) (any, error) {
	key := fmt.Sprintf("%s.%s", extensionName, servicePath)

	m.mu.RLock()
	if service, exists := m.crossServices[key]; exists {
		m.mu.RUnlock()
		return service, nil
	}
	m.mu.RUnlock()

	// Fallback to direct extraction
	extensionService, err := m.GetServiceByName(extensionName)
	if err != nil {
		return nil, fmt.Errorf("extension %s not found: %v", extensionName, err)
	}

	return m.extractServiceByPath(extensionService, servicePath)
}

// RegisterCrossService manually registers a cross service
func (m *Manager) RegisterCrossService(key string, service any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.crossServices[key] = service
	logger.Debugf(nil, "Cross service registered: %s", key)
}

// extractServiceByPath extracts service by path
func (m *Manager) extractServiceByPath(service any, path string) (any, error) {
	parts := strings.Split(path, ".")
	current := service

	for _, part := range parts {
		if field, ok := types.GetServiceInterface(current, part); ok {
			current = field
		} else {
			return nil, fmt.Errorf("service path %s not found", part)
		}
	}

	return current, nil
}

// ExecuteWithCircuitBreaker executes a function
func (m *Manager) ExecuteWithCircuitBreaker(extensionName string, fn func() (any, error)) (any, error) {
	cb, ok := m.circuitBreakers[extensionName]
	if !ok {
		return nil, fmt.Errorf("circuit breaker not found for extension %s", extensionName)
	}

	result, err := cb.Execute(fn)

	// Track circuit breaker events
	if cb.State() != gobreaker.StateClosed {
		if m.metricsManager != nil {
			m.metricsManager.CircuitBreakerTripped(extensionName)
		}
	}

	return result, err
}

// trackServiceCall tracks service call metrics
func (m *Manager) trackServiceCall(serviceName, methodName string, duration time.Duration, err error) {
	if m.metricsManager != nil {
		m.metricsManager.ServiceCall(serviceName, methodName, duration, err)
	}
}

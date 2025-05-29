package grpc

import (
	"context"
	"fmt"
)

// ServiceMethod represents a simple service method definition
type ServiceMethod struct {
	Name        string
	ServiceName string
	Handler     func(ctx context.Context, req any) (any, error)
}

// MethodBuilder helps build service method definitions
type MethodBuilder struct {
	name        string
	serviceName string
	handler     func(ctx context.Context, req any) (any, error)
}

// NewMethodBuilder creates a new service method builder
func NewMethodBuilder(serviceName, methodName string) *MethodBuilder {
	return &MethodBuilder{
		name:        methodName,
		serviceName: serviceName,
	}
}

// WithHandler sets the method handler
func (b *MethodBuilder) WithHandler(handler func(ctx context.Context, req any) (any, error)) *MethodBuilder {
	b.handler = handler
	return b
}

// Build creates the service method definition
func (b *MethodBuilder) Build() *ServiceMethod {
	return &ServiceMethod{
		Name:        b.name,
		ServiceName: b.serviceName,
		Handler:     b.handler,
	}
}

// SimpleServiceRegistry manages simple service methods
type SimpleServiceRegistry struct {
	methods map[string]*ServiceMethod
}

// NewSimpleServiceRegistry creates a new simple service registry
func NewSimpleServiceRegistry() *SimpleServiceRegistry {
	return &SimpleServiceRegistry{
		methods: make(map[string]*ServiceMethod),
	}
}

// RegisterMethod registers a service method
func (r *SimpleServiceRegistry) RegisterMethod(method *ServiceMethod) {
	key := fmt.Sprintf("%s.%s", method.ServiceName, method.Name)
	r.methods[key] = method
}

// GetMethod gets a service method
func (r *SimpleServiceRegistry) GetMethod(serviceName, methodName string) (*ServiceMethod, bool) {
	key := fmt.Sprintf("%s.%s", serviceName, methodName)
	method, exists := r.methods[key]
	return method, exists
}

// ListMethods lists all methods for a service
func (r *SimpleServiceRegistry) ListMethods(serviceName string) []*ServiceMethod {
	var methods []*ServiceMethod
	prefix := serviceName + "."

	for key, method := range r.methods {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			methods = append(methods, method)
		}
	}

	return methods
}

// ListServices lists all registered services
func (r *SimpleServiceRegistry) ListServices() []string {
	serviceSet := make(map[string]bool)

	for _, method := range r.methods {
		serviceSet[method.ServiceName] = true
	}

	var services []string
	for service := range serviceSet {
		services = append(services, service)
	}

	return services
}

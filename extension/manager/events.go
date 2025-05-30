package manager

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Event management methods

// PublishEvent publishes event
func (m *Manager) PublishEvent(eventName string, data any, target ...types.EventTarget) {
	targetFlag := m.determineEventTarget(target...)

	// Track event publication
	m.trackEventPublished(eventName, targetFlag)

	// Publish to memory dispatcher
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.Publish(eventName, data)
		m.trackEventDelivered(eventName, nil) // Assume success for memory events
	}

	// Publish to message queue
	if targetFlag&types.EventTargetQueue != 0 && m.isMessagingAvailable() {
		err := m.publishToQueue(eventName, data)
		m.trackEventDelivered(eventName, err)
	}
}

// PublishEventWithRetry publishes event with retry
func (m *Manager) PublishEventWithRetry(eventName string, data any, maxRetries int, target ...types.EventTarget) {
	targetFlag := m.determineEventTarget(target...)

	// Track event publication
	m.trackEventPublished(eventName, targetFlag)

	// Retry for memory dispatcher
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.PublishWithRetry(eventName, data, maxRetries)
		m.trackEventDelivered(eventName, nil) // Assume success for memory events
	}

	// Retry for message queue
	if targetFlag&types.EventTargetQueue != 0 && m.isMessagingAvailable() {
		err := m.publishToQueueWithRetry(eventName, data, maxRetries)
		m.trackEventDelivered(eventName, err)
	}
}

// SubscribeEvent subscribes to events from specified sources
func (m *Manager) SubscribeEvent(eventName string, handler func(any), source ...types.EventTarget) {
	sourceFlag := m.determineEventTarget(source...)

	// Subscribe to memory dispatcher
	if sourceFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.Subscribe(eventName, handler)
	}

	// Subscribe to message queue
	if sourceFlag&types.EventTargetQueue != 0 && m.isMessagingAvailable() {
		m.subscribeToQueue(eventName, handler)
	}
}

// GetExtensionPublisher returns a specific extension publisher
func (m *Manager) GetExtensionPublisher(name string, publisherType reflect.Type) (any, error) {
	ext, err := m.GetExtensionByName(name)
	if err != nil {
		return nil, err
	}

	publisher := ext.GetPublisher()
	if publisher == nil {
		return nil, fmt.Errorf("extension %s does not provide a publisher", name)
	}

	pubValue := reflect.ValueOf(publisher)
	if !pubValue.Type().ConvertibleTo(publisherType) {
		return nil, fmt.Errorf("extension %s publisher type %s is not compatible with requested type %s",
			name, pubValue.Type().String(), publisherType.String())
	}

	return publisher, nil
}

// GetExtensionSubscriber returns a specific extension subscriber
func (m *Manager) GetExtensionSubscriber(name string, subscriberType reflect.Type) (any, error) {
	ext, err := m.GetExtensionByName(name)
	if err != nil {
		return nil, err
	}

	subscriber := ext.GetSubscriber()
	if subscriber == nil {
		return nil, fmt.Errorf("extension %s does not provide a subscriber", name)
	}

	subValue := reflect.ValueOf(subscriber)
	if !subValue.Type().ConvertibleTo(subscriberType) {
		return nil, fmt.Errorf("extension %s subscriber type %s is not compatible with requested type %s",
			name, subValue.Type().String(), subscriberType.String())
	}

	return subscriber, nil
}

// GetEventsMetrics returns event metrics from the metrics system
func (m *Manager) GetEventsMetrics() map[string]any {
	if m.metricsManager == nil || !m.metricsManager.IsEnabled() {
		return map[string]any{"status": "disabled"}
	}

	// Get events collection from metrics system
	collections := m.metricsManager.GetAllCollections()
	eventsCollection, exists := collections["events"]

	if !exists {
		return map[string]any{
			"status":    "no_data",
			"timestamp": time.Now(),
		}
	}

	// Process metrics from the events collection
	metrics := map[string]any{
		"status":    "active",
		"timestamp": time.Now(),
	}

	// Extract specific metrics from the collection
	var published, delivered, failed, retries int64

	for _, metric := range eventsCollection.Metrics {
		value := metric.Value.Load()
		switch metric.Name {
		case "published_total":
			if v, ok := value.(int64); ok {
				published += v
			}
		case "delivered_total":
			if v, ok := value.(int64); ok {
				delivered += v
			}
		case "delivery_errors_total":
			if v, ok := value.(int64); ok {
				failed += v
			}
		}
	}

	// Calculate success rate
	var successRate float64
	if published > 0 {
		successRate = (float64(delivered) / float64(published)) * 100.0
	}

	metrics["published"] = published
	metrics["delivered"] = delivered
	metrics["failed"] = failed
	metrics["retries"] = retries
	metrics["success_rate"] = successRate
	metrics["last_updated"] = eventsCollection.LastUpdated

	return metrics
}

// Helper methods for event management

// determineEventTarget determines which target to use
func (m *Manager) determineEventTarget(target ...types.EventTarget) types.EventTarget {
	if len(target) > 0 {
		return target[0]
	}

	// Default: use queue if available, otherwise memory
	if m.isMessagingAvailable() {
		return types.EventTargetQueue
	}
	return types.EventTargetMemory
}

// isMessagingAvailable checks if messaging is available
func (m *Manager) isMessagingAvailable() bool {
	return m.data != nil && m.data.IsMessagingAvailable()
}

// publishToQueue publishes single event to queue
func (m *Manager) publishToQueue(eventName string, data any) error {
	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	jsonData, err := json.Marshal(eventData)
	if err != nil {
		logger.Errorf(nil, "failed to serialize event: %v", err)
		return err
	}

	if err := m.PublishMessage(eventName, eventName, jsonData); err != nil {
		logger.Warnf(nil, "failed to publish event %s to queue: %v", eventName, err)
		// Fallback to memory
		m.eventDispatcher.Publish(eventName, data)
		return err
	}

	return nil
}

// publishToQueueWithRetry publishes to queue with retry
func (m *Manager) publishToQueueWithRetry(eventName string, data any, maxRetries int) error {
	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	jsonData, err := json.Marshal(eventData)
	if err != nil {
		logger.Errorf(nil, "failed to serialize event: %v", err)
		return err
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		if err := m.PublishMessage(eventName, eventName, jsonData); err == nil {
			return nil
		}
	}

	// Final fallback to memory
	logger.Warnf(nil, "fallback to memory for event: %s", eventName)
	m.eventDispatcher.PublishWithRetry(eventName, data, maxRetries)
	return fmt.Errorf("failed to publish to queue after %d retries", maxRetries)
}

// subscribeToQueue subscribes to queue events
func (m *Manager) subscribeToQueue(eventName string, handler func(any)) {
	err := m.SubscribeToMessages(eventName, func(data []byte) error {
		var eventData types.EventData
		if err := json.Unmarshal(data, &eventData); err != nil {
			logger.Errorf(nil, "failed to unmarshal event: %v", err)
			return err
		}

		handler(eventData)
		return nil
	})

	if err != nil {
		logger.Warnf(nil, "fallback to memory subscription for: %s", eventName)
		m.eventDispatcher.Subscribe(eventName, handler)
	}
}

// Metrics tracking helper methods

func (m *Manager) trackEventPublished(eventName string, target types.EventTarget) {
	if m.metricsManager != nil {
		targetStr := m.getTargetString(target)
		m.metricsManager.EventPublished(eventName, targetStr)
	}
}

func (m *Manager) trackEventDelivered(eventName string, err error) {
	if m.metricsManager != nil {
		m.metricsManager.EventDelivered(eventName, err)
	}
}

func (m *Manager) getTargetString(target types.EventTarget) string {
	switch target {
	case types.EventTargetMemory:
		return "memory"
	case types.EventTargetQueue:
		return "queue"
	case types.EventTargetAll:
		return "all"
	default:
		return "unknown"
	}
}

// Message queue delegation methods (these delegate to data layer)

// PublishMessage publishes message to queue
func (m *Manager) PublishMessage(exchange, routingKey string, body []byte) error {
	if m.data == nil {
		return fmt.Errorf("data layer not initialized")
	}
	return m.data.PublishToRabbitMQ(exchange, routingKey, body)
}

// SubscribeToMessages subscribes to queue messages
func (m *Manager) SubscribeToMessages(queue string, handler func([]byte) error) error {
	if m.data == nil {
		return fmt.Errorf("data layer not initialized")
	}
	return m.data.ConsumeFromRabbitMQ(queue, handler)
}

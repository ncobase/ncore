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
)

// PublishEvent publishes event
func (m *Manager) PublishEvent(eventName string, data any, target ...types.EventTarget) {
	targetFlag := m.determineEventTarget(target...)

	// Auto-track event publication by extracting extension name from event name
	if extensionName := m.extractExtensionFromEventName(eventName); extensionName != "" {
		m.trackEventPublished(extensionName, eventName)
	}

	// Publish to memory dispatcher
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.Publish(eventName, data)
	}

	// Publish to message queue
	if targetFlag&types.EventTargetQueue != 0 && m.isMessagingAvailable() {
		_ = m.publishToQueue(eventName, data)
	}
}

// PublishEventWithRetry publishes event with retry
func (m *Manager) PublishEventWithRetry(eventName string, data any, maxRetries int, target ...types.EventTarget) {
	targetFlag := m.determineEventTarget(target...)

	// Auto-track event publication by extracting extension name from event name
	if extensionName := m.extractExtensionFromEventName(eventName); extensionName != "" {
		m.trackEventPublished(extensionName, eventName)
	}

	// Retry for memory dispatcher
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.PublishWithRetry(eventName, data, maxRetries)
	}

	// Retry for message queue
	if targetFlag&types.EventTargetQueue != 0 && m.isMessagingAvailable() {
		_ = m.publishToQueueWithRetry(eventName, data, maxRetries)
	}
}

// SubscribeEvent subscribes to events
func (m *Manager) SubscribeEvent(eventName string, handler func(any), source ...types.EventTarget) {
	sourceFlag := m.determineEventTarget(source...)

	// Wrap handler to auto-track event reception
	wrappedHandler := func(data any) {
		if extensionName := m.extractExtensionFromEventName(eventName); extensionName != "" {
			m.trackEventReceived(extensionName, eventName)
		}
		handler(data)
	}

	// Subscribe to memory dispatcher
	if sourceFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.Subscribe(eventName, wrappedHandler)
	}

	// Subscribe to message queue
	if sourceFlag&types.EventTargetQueue != 0 && m.isMessagingAvailable() {
		m.subscribeToQueue(eventName, wrappedHandler)
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

// extractExtensionFromEventName extracts extension name from event name using simple patterns
func (m *Manager) extractExtensionFromEventName(eventName string) string {
	// Event naming patterns:
	// 1. "extension.event" -> "extension"
	// 2. "extension.module.event" -> "extension"
	// 3. "exts.extension.ready" -> "extension"
	parts := strings.Split(eventName, ".")

	if len(parts) >= 2 {
		// Handle special system events like "exts.extension.ready"
		if parts[0] == "exts" && len(parts) >= 3 {
			if m.isRegisteredExtension(parts[1]) {
				return parts[1]
			}
		}
		// Standard format: first part is extension name
		if m.isRegisteredExtension(parts[0]) {
			return parts[0]
		}
	}

	return ""
}

// isRegisteredExtension checks if the given name is a registered extension
func (m *Manager) isRegisteredExtension(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.extensions[name]
	return exists
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

// Message queue delegation methods

// PublishMessage publishes message to available queue system (RabbitMQ or Kafka)
func (m *Manager) PublishMessage(exchange, routingKey string, body []byte) error {
	if m.data == nil {
		return fmt.Errorf("data layer not initialized")
	}

	// Try RabbitMQ first (using exchange and routingKey)
	if err := m.data.PublishToRabbitMQ(exchange, routingKey, body); err != nil {
		// If RabbitMQ fails, try Kafka (using exchange as topic)
		if kafkaErr := m.data.PublishToKafka(context.Background(), exchange, []byte(routingKey), body); kafkaErr != nil {
			return fmt.Errorf("failed to publish to both RabbitMQ (%v) and Kafka (%v)", err, kafkaErr)
		}
	}
	return nil
}

// SubscribeToMessages subscribes to messages from available queue system
func (m *Manager) SubscribeToMessages(queue string, handler func([]byte) error) error {
	if m.data == nil {
		return fmt.Errorf("data layer not initialized")
	}

	// Try RabbitMQ first
	if err := m.data.ConsumeFromRabbitMQ(queue, handler); err != nil {
		// If RabbitMQ fails, try Kafka (using queue as topic and default group)
		groupID := fmt.Sprintf("ncore-extension-%s", queue)
		if kafkaErr := m.data.ConsumeFromKafka(context.Background(), queue, groupID, handler); kafkaErr != nil {
			return fmt.Errorf("failed to subscribe to both RabbitMQ (%v) and Kafka (%v)", err, kafkaErr)
		}
	}
	return nil
}

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
	// If messaging is disabled, skip all event publishing
	if !m.isMessagingEnabled() {
		return
	}

	targetFlag := m.determineEventTarget(target...)

	if extensionName := m.extractExtensionFromEventName(eventName); extensionName != "" {
		m.trackEventPublished(extensionName, eventName)
	}

	// Publish to memory dispatcher if memory target is included
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.Publish(eventName, data)
	}

	// Publish to message queue async if queue target is included and queue is available
	if targetFlag&types.EventTargetQueue != 0 && m.isQueueAvailable() {
		go m.publishToQueue(eventName, data)
	}
}

// PublishEventWithRetry publishes event with retry
func (m *Manager) PublishEventWithRetry(eventName string, data any, maxRetries int, target ...types.EventTarget) {
	// If messaging is disabled, skip all event publishing
	if !m.isMessagingEnabled() {
		return
	}

	targetFlag := m.determineEventTarget(target...)

	if extensionName := m.extractExtensionFromEventName(eventName); extensionName != "" {
		m.trackEventPublished(extensionName, eventName)
	}

	// Publish to memory dispatcher if memory target is included
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.PublishWithRetry(eventName, data, maxRetries)
	}

	// Publish to message queue async if queue target is included and queue is available
	if targetFlag&types.EventTargetQueue != 0 && m.isQueueAvailable() {
		go m.publishToQueueWithRetry(eventName, data, maxRetries)
	}
}

// SubscribeEvent subscribes to events
func (m *Manager) SubscribeEvent(eventName string, handler func(any), source ...types.EventTarget) {
	// If messaging is disabled, skip all event subscription
	if !m.isMessagingEnabled() {
		return
	}

	sourceFlag := m.determineEventTarget(source...)

	wrappedHandler := func(data any) {
		if extensionName := m.extractExtensionFromEventName(eventName); extensionName != "" {
			m.trackEventReceived(extensionName, eventName)
		}
		handler(data)
	}

	// Subscribe to memory dispatcher if memory source is included
	if sourceFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.Subscribe(eventName, wrappedHandler)
	}

	// Subscribe to message queue if queue source is included and queue is available
	if sourceFlag&types.EventTargetQueue != 0 && m.isQueueAvailable() {
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

// isMessagingEnabled checks if messaging is enabled in data layer
func (m *Manager) isMessagingEnabled() bool {
	return m.data != nil && m.data.IsMessagingEnabled()
}

// isQueueAvailable checks if external message queues are available
func (m *Manager) isQueueAvailable() bool {
	return m.data != nil && m.data.IsQueueAvailable()
}

// shouldUseMemoryFallback checks if should use memory fallback
func (m *Manager) shouldUseMemoryFallback() bool {
	return m.data != nil && m.data.ShouldUseMemoryFallback()
}

// determineEventTarget determines which target to use based on availability
func (m *Manager) determineEventTarget(target ...types.EventTarget) types.EventTarget {
	if len(target) > 0 {
		return target[0]
	}

	// If messaging is disabled, no targets
	if !m.isMessagingEnabled() {
		return 0
	}

	// If queue is available, use all targets
	if m.isQueueAvailable() {
		return types.EventTargetAll
	}

	// If queue not available but memory fallback is enabled, use memory only
	if m.shouldUseMemoryFallback() {
		return types.EventTargetMemory
	}

	// No targets available
	return 0
}

// extractExtensionFromEventName extracts extension name from event name
func (m *Manager) extractExtensionFromEventName(eventName string) string {
	parts := strings.Split(eventName, ".")

	if len(parts) >= 2 {
		if parts[0] == "exts" && len(parts) >= 3 {
			if m.isRegisteredExtension(parts[1]) {
				return parts[1]
			}
		}
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
func (m *Manager) publishToQueue(eventName string, data any) {
	if !m.isQueueAvailable() {
		return
	}

	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	jsonData, err := json.Marshal(eventData)
	if err != nil {
		logger.Errorf(nil, "Failed to serialize event: %v", err)
		return
	}

	if err := m.PublishMessage(eventName, eventName, jsonData); err != nil {
		logger.Warnf(nil, "Failed to publish event %s to queue: %v", eventName, err)
	}
}

// publishToQueueWithRetry publishes to queue with retry
func (m *Manager) publishToQueueWithRetry(eventName string, data any, maxRetries int) {
	if !m.isQueueAvailable() {
		return
	}

	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	jsonData, err := json.Marshal(eventData)
	if err != nil {
		logger.Errorf(nil, "Failed to serialize event: %v", err)
		return
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			time.Sleep(backoff)
		}

		if err := m.PublishMessage(eventName, eventName, jsonData); err == nil {
			return
		}
	}

	logger.Warnf(nil, "Failed to publish to queue after %d retries: %s", maxRetries, eventName)
}

// subscribeToQueue subscribes to queue events
func (m *Manager) subscribeToQueue(eventName string, handler func(any)) {
	if !m.isQueueAvailable() {
		// If queue not available but memory fallback is enabled, use memory
		if m.shouldUseMemoryFallback() {
			m.eventDispatcher.Subscribe(eventName, handler)
		}
		return
	}

	err := m.SubscribeToMessages(eventName, func(data []byte) error {
		var eventData types.EventData
		if err := json.Unmarshal(data, &eventData); err != nil {
			logger.Errorf(nil, "Failed to unmarshal event: %v", err)
			return err
		}

		handler(eventData)
		return nil
	})

	if err != nil {
		logger.Warnf(nil, "Failed to subscribe to queue: %s", eventName)
		// Fallback to memory if enabled
		if m.shouldUseMemoryFallback() {
			logger.Infof(nil, "Falling back to memory subscription for: %s", eventName)
			m.eventDispatcher.Subscribe(eventName, handler)
		}
	}
}

// PublishMessage publishes message to available queue system
func (m *Manager) PublishMessage(exchange, routingKey string, body []byte) error {
	if m.data == nil {
		return fmt.Errorf("data layer not initialized")
	}

	if !m.data.IsMessagingEnabled() {
		return fmt.Errorf("messaging is disabled")
	}

	// Try RabbitMQ first
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

	if !m.data.IsMessagingEnabled() {
		return fmt.Errorf("messaging is disabled")
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

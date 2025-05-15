package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// PublishMessage publishes a message to RabbitMQ or Kafka
func (m *Manager) PublishMessage(exchange, routingKey string, body []byte) error {
	if m.data == nil {
		return fmt.Errorf("data layer not initialized")
	}

	if m.data.RabbitMQ != nil && m.data.RabbitMQ.IsConnected() {
		return m.data.PublishToRabbitMQ(exchange, routingKey, body)
	} else if m.data.Kafka != nil && m.data.Kafka.IsConnected() {
		return m.data.PublishToKafka(context.Background(), routingKey, nil, body)
	}
	return fmt.Errorf("no message queue service available")
}

// SubscribeToMessages subscribes to messages from RabbitMQ or Kafka
func (m *Manager) SubscribeToMessages(queue string, handler func([]byte) error) error {
	if m.data == nil {
		return fmt.Errorf("data layer not initialized")
	}

	if m.data.RabbitMQ != nil && m.data.RabbitMQ.IsConnected() {
		logger.Debugf(context.Background(), "Subscribing to RabbitMQ queue: %s", queue)
		return m.data.ConsumeFromRabbitMQ(queue, handler)
	} else if m.data.Kafka != nil && m.data.Kafka.IsConnected() {
		logger.Debugf(context.Background(), "Subscribing to Kafka topic: %s", queue)
		return m.data.ConsumeFromKafka(context.Background(), queue, "group", handler)
	}
	return fmt.Errorf("no message queue service available")
}

// PublishEvent publishes an event to specified targets
func (m *Manager) PublishEvent(eventName string, data any, target ...types.EventTarget) {
	// Determine target based on availability
	targetFlag := types.EventTargetMemory
	mqAvailable := m.data != nil && m.data.IsMessagingAvailable()

	if mqAvailable {
		targetFlag = types.EventTargetQueue
	}

	// Override with explicit target if provided
	if len(target) > 0 {
		targetFlag = target[0]
	}

	// Publish to in-memory event bus if requested
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventBus.Publish(eventName, data)
	}

	// Publish to message queue if available and requested
	if targetFlag&types.EventTargetQueue != 0 && mqAvailable {
		eventData := types.EventData{
			Time:      time.Now(),
			Source:    "extension",
			EventType: eventName,
			Data:      data,
		}

		jsonData, err := json.Marshal(eventData)
		if err != nil {
			logger.Errorf(context.Background(), "failed to serialize event: %v", err)
			return
		}

		logger.Debugf(context.Background(), "Publishing event %s to message queue", eventName)
		if err := m.PublishMessage(eventName, eventName, jsonData); err != nil {
			logger.Warnf(context.Background(), "failed to publish event %s to message queue: %v", eventName, err)

			// Fallback to in-memory if queue publishing fails and not already published to memory
			if targetFlag&types.EventTargetMemory == 0 {
				logger.Infof(context.Background(), "falling back to in-memory event bus for event: %s", eventName)
				m.eventBus.Publish(eventName, data)
			}
		} else {
			logger.Debugf(context.Background(), "Successfully published event %s to message queue", eventName)
		}
	}
}

// PublishEventWithRetry publishes an event with retry logic
func (m *Manager) PublishEventWithRetry(eventName string, data any, maxRetries int, target ...types.EventTarget) {
	// Determine target based on availability
	targetFlag := types.EventTargetMemory
	mqAvailable := m.data != nil && m.data.IsMessagingAvailable()

	if mqAvailable {
		targetFlag = types.EventTargetQueue
	}

	// Override with explicit target if provided
	if len(target) > 0 {
		targetFlag = target[0]
	}

	// Publish to in-memory event bus with retry if requested
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventBus.PublishWithRetry(eventName, data, maxRetries)
	}

	// Publish to message queue with retry if available and requested
	if targetFlag&types.EventTargetQueue != 0 && mqAvailable {
		eventData := types.EventData{
			Time:      time.Now(),
			Source:    "extension",
			EventType: eventName,
			Data:      data,
		}

		jsonData, err := json.Marshal(eventData)
		if err != nil {
			logger.Errorf(context.Background(), "failed to serialize event: %v", err)
			return
		}

		var attempts int
		var publishErr error
		for attempts <= maxRetries {
			logger.Debugf(context.Background(), "Attempting to publish event %s to message queue (attempt %d)", eventName, attempts+1)
			publishErr = m.PublishMessage(eventName, eventName, jsonData)
			if publishErr == nil {
				logger.Debugf(context.Background(), "Successfully published event %s to message queue on attempt %d", eventName, attempts+1)
				return
			}

			logger.Warnf(context.Background(), "Failed to publish event %s (attempt %d): %v", eventName, attempts+1, publishErr)
			attempts++

			if attempts <= maxRetries {
				backoff := time.Duration(attempts) * time.Second
				logger.Debugf(context.Background(), "Retrying in %v...", backoff)
				time.Sleep(backoff)
			}
		}

		logger.Errorf(context.Background(), "Failed to publish event %s to message queue after %d retries: %v", eventName, maxRetries, publishErr)

		// Fallback to in-memory if queue publishing fails and not already published to memory
		if targetFlag&types.EventTargetMemory == 0 {
			logger.Infof(context.Background(), "falling back to in-memory event bus for event: %s", eventName)
			m.eventBus.PublishWithRetry(eventName, data, maxRetries)
		}
	}
}

// SubscribeEvent registers a handler for events from in-memory bus and/or message queue
func (m *Manager) SubscribeEvent(eventName string, handler func(any), source ...types.EventTarget) {
	// Determine source based on availability
	sourceFlag := types.EventTargetMemory
	mqAvailable := m.data != nil && m.data.IsMessagingAvailable()

	if mqAvailable {
		sourceFlag = types.EventTargetQueue
	}

	// Override with explicit source if provided
	if len(source) > 0 {
		sourceFlag = source[0]
	}

	// Subscribe to in-memory event bus if requested or as fallback
	if sourceFlag&types.EventTargetMemory != 0 {
		logger.Debugf(context.Background(), "Subscribing to in-memory event bus for event: %s", eventName)
		m.eventBus.Subscribe(eventName, handler)
	}

	// Subscribe to message queue if available and requested
	if sourceFlag&types.EventTargetQueue != 0 && mqAvailable {
		logger.Debugf(context.Background(), "Subscribing to message queue for event: %s", eventName)

		err := m.SubscribeToMessages(eventName, func(data []byte) error {
			var eventData types.EventData
			if err := json.Unmarshal(data, &eventData); err != nil {
				logger.Errorf(context.Background(), "Failed to unmarshal event data: %v", err)
				return err
			}

			logger.Debugf(context.Background(), "Received event %s from message queue", eventName)
			handler(eventData)
			return nil
		})

		if err != nil {
			logger.Warnf(context.Background(), "Failed to subscribe to message queue for event %s: %v", eventName, err)

			// If we haven't already subscribed to in-memory bus, do it as fallback
			if sourceFlag&types.EventTargetMemory == 0 {
				logger.Infof(context.Background(), "Falling back to in-memory event bus for event: %s", eventName)
				m.eventBus.Subscribe(eventName, handler)
			}
		} else {
			logger.Debugf(context.Background(), "Successfully subscribed to message queue for event: %s", eventName)
		}
	}
}

// GetEventBusMetrics returns event bus metrics
func (m *Manager) GetEventBusMetrics() map[string]any {
	return m.eventBus.GetMetrics()
}

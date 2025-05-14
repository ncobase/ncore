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
	if m.data.RabbitMQ != nil {
		return m.data.PublishToRabbitMQ(exchange, routingKey, body)
	} else if m.data.Kafka != nil {
		return m.data.PublishToKafka(context.Background(), routingKey, nil, body)
	}
	return fmt.Errorf("no message queue service available")
}

// SubscribeToMessages subscribes to messages from RabbitMQ or Kafka
func (m *Manager) SubscribeToMessages(queue string, handler func([]byte) error) error {
	if m.data.RabbitMQ != nil {
		return m.data.ConsumeFromRabbitMQ(queue, handler)
	} else if m.data.Kafka != nil {
		return m.data.ConsumeFromKafka(context.Background(), queue, "group", handler)
	}
	return fmt.Errorf("no message queue service available")
}

// PublishEvent publishes an event to specified targets
// Default: publish to in-memory event bus (safest option)
func (m *Manager) PublishEvent(eventName string, data any, target ...types.EventTarget) {
	// Default to in-memory for safety
	targetFlag := types.EventTargetMemory

	// Only use message queue if explicitly requested AND available
	if len(target) > 0 {
		targetFlag = target[0]
	}

	// Publish to in-memory event bus if requested (default behavior)
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventBus.Publish(eventName, data)
	}

	// Publish to message queue if available and requested
	if targetFlag&types.EventTargetQueue != 0 && (m.data.RabbitMQ != nil || m.data.Kafka != nil) {
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

		if err := m.PublishMessage(eventName, eventName, jsonData); err != nil {
			logger.Errorf(context.Background(), "failed to publish event to message queue: %v", err)
		}
	}
}

// PublishEventWithRetry publishes an event with retry logic
func (m *Manager) PublishEventWithRetry(eventName string, data any, maxRetries int, target ...types.EventTarget) {
	// Default to in-memory for safety
	targetFlag := types.EventTargetMemory

	// Only use message queue if explicitly requested AND available
	if len(target) > 0 {
		targetFlag = target[0]
	}

	// Publish to in-memory event bus with retry if requested (default behavior)
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventBus.PublishWithRetry(eventName, data, maxRetries)
	}

	// Publish to message queue with retry if available and requested
	if targetFlag&types.EventTargetQueue != 0 && (m.data.RabbitMQ != nil || m.data.Kafka != nil) {
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
		for attempts < maxRetries {
			if err := m.PublishMessage(eventName, eventName, jsonData); err == nil {
				return
			}
			attempts++
			time.Sleep(time.Duration(attempts) * time.Second)
		}
		logger.Errorf(context.Background(), "failed to publish event to message queue after %d retries", maxRetries)
	}
}

// SubscribeEvent registers a handler for events from in-memory bus and/or message queue
// Default: subscribe to in-memory event bus (safest option)
func (m *Manager) SubscribeEvent(eventName string, handler func(any), source ...types.EventTarget) {
	// Default to in-memory for safety
	sourceFlag := types.EventTargetMemory

	// Only use message queue if explicitly requested AND available
	if len(source) > 0 {
		sourceFlag = source[0]
	}

	// Subscribe to in-memory event bus if requested (default behavior)
	if sourceFlag&types.EventTargetMemory != 0 {
		m.eventBus.Subscribe(eventName, handler)
	}

	// Subscribe to message queue if available and requested
	if sourceFlag&types.EventTargetQueue != 0 && (m.data.RabbitMQ != nil || m.data.Kafka != nil) {
		err := m.SubscribeToMessages(eventName, func(data []byte) error {
			var eventData types.EventData
			if err := json.Unmarshal(data, &eventData); err != nil {
				return err
			}
			handler(eventData)
			return nil
		})

		if err != nil {
			logger.Warnf(context.Background(), "failed to subscribe to message queue events: %v", err)
		}
	}
}

// GetEventBusMetrics returns event bus metrics
func (m *Manager) GetEventBusMetrics() map[string]any {
	return m.eventBus.GetMetrics()
}

package manager

import (
	"context"
	"fmt"
)

// PublishMessage publishes a message using RabbitMQ or Kafka
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

// PublishEvent publishes an event to all extensions
func (m *Manager) PublishEvent(eventName string, data any) {
	m.eventBus.Publish(eventName, data)
}

// SubscribeEvent allows a extension to subscribe to an event
func (m *Manager) SubscribeEvent(eventName string, handler func(any)) {
	m.eventBus.Subscribe(eventName, handler)
}

// GetEventBusMetrics returns event bus metrics
func (m *Manager) GetEventBusMetrics() map[string]any {
	return m.eventBus.GetMetrics()
}

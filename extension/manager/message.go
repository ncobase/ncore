package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// MessageQueueMetrics tracks message queue statistics
type MessageQueueMetrics struct {
	published       atomic.Int64
	publishFailed   atomic.Int64
	consumed        atomic.Int64
	consumeFailed   atomic.Int64
	retryAttempts   atomic.Int64
	lastPublishTime atomic.Value // time.Time
	lastConsumeTime atomic.Value // time.Time
}

// NewMessageQueueMetrics creates new message queue metrics
func NewMessageQueueMetrics() *MessageQueueMetrics {
	m := &MessageQueueMetrics{}
	m.lastPublishTime.Store(time.Time{})
	m.lastConsumeTime.Store(time.Time{})
	return m
}

// PublishMessage publishes message to queue with metrics
func (m *Manager) PublishMessage(exchange, routingKey string, body []byte) error {
	if m.data == nil {
		m.mqMetrics.publishFailed.Add(1)
		return fmt.Errorf("data layer not initialized")
	}

	var err error
	if m.data.RabbitMQ != nil && m.data.RabbitMQ.IsConnected() {
		err = m.data.PublishToRabbitMQ(exchange, routingKey, body)
	} else if m.data.Kafka != nil && m.data.Kafka.IsConnected() {
		err = m.data.PublishToKafka(context.Background(), routingKey, nil, body)
	} else {
		err = fmt.Errorf("no message queue service available")
	}

	if err != nil {
		m.mqMetrics.publishFailed.Add(1)
	} else {
		m.mqMetrics.published.Add(1)
		m.mqMetrics.lastPublishTime.Store(time.Now())
	}

	return err
}

// SubscribeToMessages subscribes to queue messages with metrics
func (m *Manager) SubscribeToMessages(queue string, handler func([]byte) error) error {
	if m.data == nil {
		return fmt.Errorf("data layer not initialized")
	}

	// Wrap handler with metrics
	wrappedHandler := func(data []byte) error {
		m.mqMetrics.lastConsumeTime.Store(time.Now())

		err := handler(data)
		if err != nil {
			m.mqMetrics.consumeFailed.Add(1)
		} else {
			m.mqMetrics.consumed.Add(1)
		}
		return err
	}

	if m.data.RabbitMQ != nil && m.data.RabbitMQ.IsConnected() {
		return m.data.ConsumeFromRabbitMQ(queue, wrappedHandler)
	} else if m.data.Kafka != nil && m.data.Kafka.IsConnected() {
		return m.data.ConsumeFromKafka(context.Background(), queue, "group", wrappedHandler)
	}

	return fmt.Errorf("no message queue service available")
}

// PublishEvent publishes event to specified targets with unified metrics
func (m *Manager) PublishEvent(eventName string, data any, target ...types.EventTarget) {
	targetFlag := m.determineEventTarget(target...)

	// Publish to memory dispatcher
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.Publish(eventName, data)
	}

	// Publish to message queue
	if targetFlag&types.EventTargetQueue != 0 && m.isMessagingAvailable() {
		m.publishToQueue(eventName, data)
	}
}

// PublishEventWithRetry publishes event with retry logic and metrics
func (m *Manager) PublishEventWithRetry(eventName string, data any, maxRetries int, target ...types.EventTarget) {
	targetFlag := m.determineEventTarget(target...)

	// Retry for memory dispatcher
	if targetFlag&types.EventTargetMemory != 0 {
		m.eventDispatcher.PublishWithRetry(eventName, data, maxRetries)
	}

	// Retry for message queue
	if targetFlag&types.EventTargetQueue != 0 && m.isMessagingAvailable() {
		m.publishToQueueWithRetry(eventName, data, maxRetries)
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
func (m *Manager) publishToQueue(eventName string, data any) {
	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	jsonData, err := json.Marshal(eventData)
	if err != nil {
		logger.Errorf(nil, "failed to serialize event: %v", err)
		return
	}

	if err := m.PublishMessage(eventName, eventName, jsonData); err != nil {
		logger.Warnf(nil, "failed to publish event %s to queue: %v", eventName, err)
		// Fallback to memory
		m.eventDispatcher.Publish(eventName, data)
	}
}

// publishToQueueWithRetry publishes to queue with retry
func (m *Manager) publishToQueueWithRetry(eventName string, data any, maxRetries int) {
	eventData := types.EventData{
		Time:      time.Now(),
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	jsonData, err := json.Marshal(eventData)
	if err != nil {
		logger.Errorf(nil, "failed to serialize event: %v", err)
		return
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		m.mqMetrics.retryAttempts.Add(1)

		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		if err := m.PublishMessage(eventName, eventName, jsonData); err == nil {
			return
		}
	}

	// Final fallback to memory
	logger.Warnf(nil, "fallback to memory for event: %s", eventName)
	m.eventDispatcher.PublishWithRetry(eventName, data, maxRetries)
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

// GetEventsMetrics returns comprehensive event metrics
func (m *Manager) GetEventsMetrics() map[string]any {
	if m.eventDispatcher == nil {
		return map[string]any{"status": "not_initialized"}
	}

	memoryMetrics := m.eventDispatcher.GetMetrics()
	queueMetrics := m.getQueueMetrics()

	// Calculate totals
	totalPublished := memoryMetrics["published"].(int64) + queueMetrics["published"].(int64)
	totalFailed := memoryMetrics["failed"].(int64) + queueMetrics["failed"].(int64)
	totalSuccess := totalPublished - totalFailed

	var overallSuccessRate float64
	if totalPublished > 0 {
		overallSuccessRate = (float64(totalSuccess) / float64(totalPublished)) * 100.0
	}

	return map[string]any{
		"status": "active",
		"memory": memoryMetrics,
		"queue":  queueMetrics,
		"total": map[string]any{
			"published":    totalPublished,
			"failed":       totalFailed,
			"success":      totalSuccess,
			"success_rate": overallSuccessRate,
		},
		"timestamp": time.Now(),
	}
}

// getQueueMetrics returns message queue metrics
func (m *Manager) getQueueMetrics() map[string]any {
	if m.mqMetrics == nil {
		return map[string]any{
			"published":      int64(0),
			"failed":         int64(0),
			"consumed":       int64(0),
			"consume_failed": int64(0),
			"retries":        int64(0),
		}
	}

	published := m.mqMetrics.published.Load()
	publishFailed := m.mqMetrics.publishFailed.Load()
	consumed := m.mqMetrics.consumed.Load()
	consumeFailed := m.mqMetrics.consumeFailed.Load()

	var publishSuccessRate float64
	if published+publishFailed > 0 {
		publishSuccessRate = (float64(published) / float64(published+publishFailed)) * 100.0
	}

	var consumeSuccessRate float64
	if consumed+consumeFailed > 0 {
		consumeSuccessRate = (float64(consumed) / float64(consumed+consumeFailed)) * 100.0
	}

	return map[string]any{
		"published":            published,
		"failed":               publishFailed,
		"consumed":             consumed,
		"consume_failed":       consumeFailed,
		"retries":              m.mqMetrics.retryAttempts.Load(),
		"publish_success_rate": publishSuccessRate,
		"consume_success_rate": consumeSuccessRate,
		"last_publish_time":    m.mqMetrics.lastPublishTime.Load(),
		"last_consume_time":    m.mqMetrics.lastConsumeTime.Load(),
	}
}

// getMessagingHealthStatus returns messaging system health
func (m *Manager) getMessagingHealthStatus() map[string]any {
	if m.data == nil {
		return map[string]any{
			"status":          "unavailable",
			"reason":          "data layer not initialized",
			"memory_fallback": true,
			"fallback_active": m.eventDispatcher != nil,
		}
	}

	rabbitmqConnected := false
	kafkaConnected := false

	if m.data.RabbitMQ != nil {
		rabbitmqConnected = m.data.RabbitMQ.IsConnected()
	}

	if m.data.Kafka != nil {
		kafkaConnected = m.data.Kafka.IsConnected()
	}

	overallAvailable := m.data.IsMessagingAvailable()
	memoryFallbackActive := !overallAvailable && m.eventDispatcher != nil

	status := map[string]any{
		"rabbitmq_connected": rabbitmqConnected,
		"kafka_connected":    kafkaConnected,
		"overall_available":  overallAvailable,
		"memory_fallback":    true,
		"fallback_active":    memoryFallbackActive,
		"primary_transport":  m.getPrimaryTransport(overallAvailable),
		"fallback_reason":    m.getFallbackReason(rabbitmqConnected, kafkaConnected),
	}

	if memoryFallbackActive {
		status["fallback_metrics"] = m.eventDispatcher.GetMetrics()
	}

	return status
}

// getPrimaryTransport returns the current primary transport method
func (m *Manager) getPrimaryTransport(queueAvailable bool) string {
	if queueAvailable {
		if m.data.RabbitMQ != nil && m.data.RabbitMQ.IsConnected() {
			return "rabbitmq"
		}
		if m.data.Kafka != nil && m.data.Kafka.IsConnected() {
			return "kafka"
		}
	}
	return "memory"
}

// getFallbackReason returns reason for fallback activation
func (m *Manager) getFallbackReason(rabbitmqConnected, kafkaConnected bool) string {
	if !rabbitmqConnected && !kafkaConnected {
		return "no_queue_services_available"
	}
	if !rabbitmqConnected && m.data.RabbitMQ != nil {
		return "rabbitmq_disconnected"
	}
	if !kafkaConnected && m.data.Kafka != nil {
		return "kafka_disconnected"
	}
	return "none"
}

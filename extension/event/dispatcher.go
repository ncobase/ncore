package event

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Dispatcher handles event publishing and subscription
type Dispatcher struct {
	subscribers map[string][]func(any)
	mu          sync.RWMutex
	metrics     struct {
		published        atomic.Int64
		delivered        atomic.Int64
		failed           atomic.Int64
		retries          atomic.Int64
		lastEventTime    atomic.Value // time.Time
		activeHandlers   atomic.Int32
		totalSubscribers atomic.Int32
		startTime        time.Time
		recentEvents     []time.Time
		recentEventsMu   sync.Mutex
	}
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *Dispatcher {
	d := &Dispatcher{
		subscribers: make(map[string][]func(any)),
	}
	d.metrics.lastEventTime.Store(time.Time{})
	d.metrics.startTime = time.Now()
	d.metrics.recentEvents = make([]time.Time, 0, 1000)
	return d
}

// Subscribe adds a handler for specific event
func (d *Dispatcher) Subscribe(eventName string, handler func(any)) {
	if handler == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	wrappedHandler := d.wrapHandler(handler)
	d.subscribers[eventName] = append(d.subscribers[eventName], wrappedHandler)
	d.metrics.totalSubscribers.Add(1)
}

// Publish sends event to all subscribers
func (d *Dispatcher) Publish(eventName string, data any) {
	d.mu.RLock()
	handlers, exists := d.subscribers[eventName]
	handlerCount := len(handlers)
	d.mu.RUnlock()

	if !exists || handlerCount == 0 {
		return
	}

	// Count actual deliveries to handlers
	d.metrics.published.Add(int64(handlerCount))

	now := time.Now()
	d.metrics.lastEventTime.Store(now)
	d.recordRecentEvent(now)

	eventData := types.EventData{
		Time:      now,
		Source:    "extension",
		EventType: eventName,
		Data:      data,
	}

	// Execute handlers concurrently
	for _, handler := range handlers {
		go handler(eventData)
	}
}

// PublishWithRetry publishes event with retry
func (d *Dispatcher) PublishWithRetry(eventName string, data any, maxRetries int) {
	d.mu.RLock()
	handlers, exists := d.subscribers[eventName]
	handlerCount := len(handlers)
	d.mu.RUnlock()

	if !exists || handlerCount == 0 {
		d.metrics.failed.Add(1)
		return
	}

	success := false
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			d.metrics.retries.Add(1)
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		d.Publish(eventName, data)
		success = true
		break
	}

	if !success {
		d.metrics.failed.Add(int64(handlerCount))
	}
}

// recordRecentEvent records event time for rate calculation
func (d *Dispatcher) recordRecentEvent(eventTime time.Time) {
	d.metrics.recentEventsMu.Lock()
	defer d.metrics.recentEventsMu.Unlock()

	d.metrics.recentEvents = append(d.metrics.recentEvents, eventTime)

	// Clean up events older than 60 seconds
	cutoff := eventTime.Add(-60 * time.Second)
	start := 0
	for i, t := range d.metrics.recentEvents {
		if t.After(cutoff) {
			start = i
			break
		}
	}

	if start > 0 {
		copy(d.metrics.recentEvents, d.metrics.recentEvents[start:])
		d.metrics.recentEvents = d.metrics.recentEvents[:len(d.metrics.recentEvents)-start]
	}

	// Limit max length to prevent memory leak
	if len(d.metrics.recentEvents) > 1000 {
		start = len(d.metrics.recentEvents) - 1000
		copy(d.metrics.recentEvents, d.metrics.recentEvents[start:])
		d.metrics.recentEvents = d.metrics.recentEvents[:1000]
	}
}

// wrapHandler wraps user handler with metrics
func (d *Dispatcher) wrapHandler(handler func(any)) func(any) {
	return func(data any) {
		d.metrics.activeHandlers.Add(1)
		defer d.metrics.activeHandlers.Add(-1)

		defer func() {
			if r := recover(); r != nil {
				d.metrics.failed.Add(1)
				logger.Errorf(nil, "event handler panic: %v", r)
				return
			}
			d.metrics.delivered.Add(1)
		}()

		handler(data)
	}
}

// GetMetrics returns comprehensive metrics
func (d *Dispatcher) GetMetrics() map[string]any {
	lastEventTime := d.metrics.lastEventTime.Load().(time.Time)
	published := d.metrics.published.Load()
	delivered := d.metrics.delivered.Load()
	failed := d.metrics.failed.Load()

	var successRate float64
	if published > 0 {
		successRate = (float64(delivered) / float64(published)) * 100.0
	}

	uptime := time.Since(d.metrics.startTime)

	return map[string]any{
		"published":         published,
		"delivered":         delivered,
		"failed":            failed,
		"retries":           d.metrics.retries.Load(),
		"success_rate":      successRate,
		"last_event_time":   lastEventTime,
		"active_handlers":   d.metrics.activeHandlers.Load(),
		"total_subscribers": d.metrics.totalSubscribers.Load(),
		"events_per_second": d.calculateEventsPerSecond(),
		"events_per_minute": d.calculateEventsPerMinute(),
		"uptime_seconds":    uptime.Seconds(),
		"average_rate":      d.calculateAverageRate(published, uptime),
		"recent_activity":   d.getRecentActivity(),
	}
}

// calculateEventsPerSecond calculates events rate in last 10 seconds
func (d *Dispatcher) calculateEventsPerSecond() float64 {
	d.metrics.recentEventsMu.Lock()
	defer d.metrics.recentEventsMu.Unlock()

	if len(d.metrics.recentEvents) == 0 {
		return 0.0
	}

	now := time.Now()
	cutoff := now.Add(-10 * time.Second)
	count := 0

	for _, eventTime := range d.metrics.recentEvents {
		if eventTime.After(cutoff) {
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return float64(count) / 10.0
}

// calculateEventsPerMinute calculates events rate in last 60 seconds
func (d *Dispatcher) calculateEventsPerMinute() float64 {
	d.metrics.recentEventsMu.Lock()
	defer d.metrics.recentEventsMu.Unlock()

	if len(d.metrics.recentEvents) == 0 {
		return 0.0
	}

	now := time.Now()
	cutoff := now.Add(-60 * time.Second)
	count := 0

	for _, eventTime := range d.metrics.recentEvents {
		if eventTime.After(cutoff) {
			count++
		}
	}

	return float64(count)
}

// calculateAverageRate calculates overall average event rate
func (d *Dispatcher) calculateAverageRate(published int64, uptime time.Duration) float64 {
	if uptime.Seconds() <= 0 || published == 0 {
		return 0.0
	}

	return float64(published) / uptime.Seconds()
}

// getRecentActivity returns recent activity status
func (d *Dispatcher) getRecentActivity() map[string]any {
	lastEventTime := d.metrics.lastEventTime.Load().(time.Time)

	var status string
	var timeSinceLastEvent float64

	if lastEventTime.IsZero() {
		status = "no_events"
		timeSinceLastEvent = 0
	} else {
		timeSinceLastEvent = time.Since(lastEventTime).Seconds()
		switch {
		case timeSinceLastEvent < 10:
			status = "very_active"
		case timeSinceLastEvent < 60:
			status = "active"
		case timeSinceLastEvent < 300:
			status = "idle"
		default:
			status = "inactive"
		}
	}

	return map[string]any{
		"status":                status,
		"seconds_since_last":    timeSinceLastEvent,
		"last_event_time":       lastEventTime,
		"recent_events_tracked": len(d.metrics.recentEvents),
	}
}

package metrics

import "time"

// ExtensionCollectorAdapter adapts extension metrics collector to data layer interface
type ExtensionCollectorAdapter struct {
	collector ExtensionCollector
}

// ExtensionCollector interface from extension layer
type ExtensionCollector interface {
	DBQuery(duration time.Duration, err error)
	DBTransaction(err error)
	DBConnections(count int)
	RedisCommand(command string, err error)
	RedisConnections(count int)
	MongoOperation(operation string, err error)
	SearchQuery(engine string, err error)
	SearchIndex(engine, operation string)
	MQPublish(system string, err error)
	MQConsume(system string, err error)
	HealthCheck(component string, healthy bool)
}

// NewExtensionCollectorAdapter creates adapter for extension collector
func NewExtensionCollectorAdapter(collector ExtensionCollector) *ExtensionCollectorAdapter {
	return &ExtensionCollectorAdapter{collector: collector}
}

// Implement data layer Collector interface
func (a *ExtensionCollectorAdapter) DBQuery(duration time.Duration, err error) {
	if a.collector != nil {
		a.collector.DBQuery(duration, err)
	}
}

func (a *ExtensionCollectorAdapter) DBTransaction(err error) {
	if a.collector != nil {
		a.collector.DBTransaction(err)
	}
}

func (a *ExtensionCollectorAdapter) DBConnections(count int) {
	if a.collector != nil {
		a.collector.DBConnections(count)
	}
}

func (a *ExtensionCollectorAdapter) RedisCommand(command string, err error) {
	if a.collector != nil {
		a.collector.RedisCommand(command, err)
	}
}

func (a *ExtensionCollectorAdapter) RedisConnections(count int) {
	if a.collector != nil {
		a.collector.RedisConnections(count)
	}
}

func (a *ExtensionCollectorAdapter) MongoOperation(operation string, err error) {
	if a.collector != nil {
		a.collector.MongoOperation(operation, err)
	}
}

func (a *ExtensionCollectorAdapter) SearchQuery(engine string, err error) {
	if a.collector != nil {
		a.collector.SearchQuery(engine, err)
	}
}

func (a *ExtensionCollectorAdapter) SearchIndex(engine, operation string) {
	if a.collector != nil {
		a.collector.SearchIndex(engine, operation)
	}
}

func (a *ExtensionCollectorAdapter) MQPublish(system string, err error) {
	if a.collector != nil {
		a.collector.MQPublish(system, err)
	}
}

func (a *ExtensionCollectorAdapter) MQConsume(system string, err error) {
	if a.collector != nil {
		a.collector.MQConsume(system, err)
	}
}

func (a *ExtensionCollectorAdapter) HealthCheck(component string, healthy bool) {
	if a.collector != nil {
		a.collector.HealthCheck(component, healthy)
	}
}

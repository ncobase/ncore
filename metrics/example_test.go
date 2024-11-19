package metrics_test

//
// import (
// 	"context"
// 	"fmt"
// 	"ncobase/common/metrics"
// 	"time"
// )
//
// func ExampleCollector() {
// 	// Create a new metrics collector with default configuration
// 	collector, err := metrics.NewCollector(metrics.DefaultConfig())
// 	if err != nil {
// 		fmt.Printf("Failed to create collector: %v\n", err)
// 		return
// 	}
//
// 	// Start the collector
// 	ctx := context.Background()
// 	if err := collector.Start(ctx); err != nil {
// 		fmt.Printf("Failed to start collector: %v\n", err)
// 		return
// 	}
// 	defer collector.Stop()
//
// 	// Record some metrics
// 	collector.RecordProcessStart()
// 	time.Sleep(time.Second)
// 	collector.RecordProcessCompletion(time.Since(time.Now()).Seconds(), true)
//
// 	collector.RecordExecutorAttemptDuration("task_executor", time.Second, nil)
// 	collector.RecordExecutorRetryCount("task_executor", 1, fmt.Errorf("temporary failure"))
//
// 	// Print the collected metrics
// 	fmt.Printf("Metrics: %+v\n", collector.GetMetrics())
//
// 	// Output:
// 	// Metrics: map[process:map[active:0 completed:1 durations:map[count:1 max:1 mean:1 min:1 percentiles:map[50:1 75:1 90:1 95:1 99:1]] failed:0 total:1] runtime:map[start_time:1621234567 uptime:1]]
// }
//
// func Example_collectorRecordTaskStart() {
// 	collector, _ := metrics.NewCollector(metrics.DefaultConfig())
// 	ctx := context.Background()
// 	collector.Start(ctx)
// 	defer collector.Stop()
//
// 	collector.RecordTaskStart()
// 	// Output:
// }
//
// func Example_collectorRecordTaskCompletion() {
// 	collector, _ := metrics.NewCollector(metrics.DefaultConfig())
// 	ctx := context.Background()
// 	collector.Start(ctx)
// 	defer collector.Stop()
//
// 	collector.RecordTaskStart()
// 	time.Sleep(time.Second)
// 	collector.RecordTaskCompletion(time.Since(time.Now()).Seconds(), false)
// 	// Output:
// }
//
// func Example_collectorRecordNodeExecution() {
// 	collector, _ := metrics.NewCollector(metrics.DefaultConfig())
// 	ctx := context.Background()
// 	collector.Start(ctx)
// 	defer collector.Stop()
//
// 	collector.Node().RecordNodeExecution("activity", time.Second.Seconds(), true, false)
// 	// Output:
// }
//
// func Example_collectorRecordHandlerExecutionDuration() {
// 	collector, _ := metrics.NewCollector(metrics.DefaultConfig())
// 	ctx := context.Background()
// 	collector.Start(ctx)
// 	defer collector.Stop()
//
// 	collector.RecordHandlerExecutionDuration("decision_handler", time.Second, nil)
// 	// Output:
// }
//
// func Example_collectorRecordRetry() {
// 	collector, _ := metrics.NewCollector(metrics.DefaultConfig())
// 	ctx := context.Background()
// 	collector.Start(ctx)
// 	defer collector.Stop()
//
// 	event := metrics.RetryEvent{
// 		Attempt:  1,
// 		Duration: time.Second,
// 		Error:    fmt.Errorf("temporary error"),
// 	}
// 	collector.RecordRetry(event)
// 	// Output:
// }

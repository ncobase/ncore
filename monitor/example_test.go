package monitor_test

//
// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"ncobase/ncore/monitor"
// 	"net/http"
// 	"time"
// )
//
// // ExampleServiceMonitoring demonstrates how to use RuntimeStats for service monitoring
// // and automatic protection mechanisms.
// func Example_serviceMonitoring() {
// 	ctx := context.Background()
// 	config := &monitor.RuntimeStatsConfig{
// 		MaxMemory:     2 * 1024 * 1024 * 1024, // 2GB
// 		MaxCPU:        80,                     // 80%
// 		MaxGoroutines: 5000,                   // 5k goroutines
// 		Interval:      time.Second,
// 	}
//
// 	m, err := monitor.NewRuntimeStats(ctx, config)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	m.Start()
// 	defer m.Stop()
//
// 	// Simulated monitoring loop
// 	tick := time.NewTicker(time.Second)
// 	select {
// 	case <-tick.C:
// 		if errs := m.CheckThresholds(); len(errs) > 0 {
// 			// Trigger alerts and protection mechanisms
// 			fmt.Println("Runtime metrics exceeded thresholds")
// 			for _, err := range errs {
// 				fmt.Printf("Alert: %v\n", err)
// 			}
// 		}
// 	case <-ctx.Done():
// 		return
// 	}
//
// 	// Output:
// 	// Runtime metrics exceeded thresholds
// 	// Alert: memory usage exceeded: 2147483648 > 2147483648
// }
//
// // ExampleHealthCheck shows how to implement a health check endpoint
// // using RuntimeStats.
// func Example_healthCheck() {
// 	type HealthStatus struct {
// 		Status     string    `json:"status"`
// 		Memory     int64     `json:"memory"`
// 		CPU        float64   `json:"cpu"`
// 		Goroutines int32     `json:"goroutines"`
// 		Errors     []error   `json:"errors,omitempty"`
// 		Timestamp  time.Time `json:"timestamp"`
// 	}
//
// 	ctx := context.Background()
// 	m, _ := monitor.NewRuntimeStats(ctx, nil)
// 	m.Start()
// 	defer m.Stop()
//
// 	// Simulate HTTP handler
// 	handler := func(w http.ResponseWriter, r *http.Request) {
// 		usage := m.GetMetrics()
// 		health := HealthStatus{
// 			Status:     "healthy",
// 			Memory:     usage.Memory,
// 			CPU:        usage.CPU,
// 			Goroutines: usage.Goroutines,
// 			Timestamp:  time.Now(),
// 		}
//
// 		if errs := m.CheckThresholds(); len(errs) > 0 {
// 			health.Status = "unhealthy"
// 			health.Errors = errs
// 		}
//
// 		// For example output
// 		out, _ := json.MarshalIndent(health, "", "  ")
// 		fmt.Println(string(out))
// 	}
//
// 	// Simulate request
// 	handler(nil, nil)
//
// 	// Output example will vary based on actual runtime usage
// }
//
// // ExampleLoadBalancer demonstrates how to use RuntimeStats for load balancing decisions.
// func Example_loadBalancer() {
// 	type ServiceInstance struct {
// 		ID      string
// 		Monitor *monitor.RuntimeStats
// 	}
//
// 	// Calculate load score based on runtime usage
// 	calculateLoadScore := func(usage monitor.Metrics) float64 {
// 		// Simple scoring example:
// 		// 50% weight to CPU, 30% to memory, 20% to goroutine count
// 		memoryScore := float64(usage.Memory) / float64(1<<30) * 100 // Convert to GB and percentage
// 		goroutineScore := float64(usage.Goroutines) / 1000 * 100    // Assume 1000 is baseline
// 		return usage.CPU*0.5 + memoryScore*0.3 + goroutineScore*0.2
// 	}
//
// 	// Create sample instances
// 	ctx := context.Background()
// 	instances := []ServiceInstance{
// 		{ID: "instance-1", Monitor: mustCreateMonitor(ctx)},
// 		{ID: "instance-2", Monitor: mustCreateMonitor(ctx)},
// 	}
//
// 	// Select instance with lowest load
// 	var selected *ServiceInstance
// 	minLoad := float64(100)
//
// 	for i := range instances {
// 		usage := instances[i].Monitor.GetMetrics()
// 		load := calculateLoadScore(usage)
// 		if load < minLoad {
// 			minLoad = load
// 			selected = &instances[i]
// 		}
// 	}
//
// 	if selected == nil {
// 		fmt.Println("No instances available")
// 		return
// 	}
//
// 	fmt.Printf("Selected instance: %s with load score: %.2f\n", selected.ID, minLoad)
//
// 	// Cleanup
// 	for _, inst := range instances {
// 		inst.Monitor.Stop()
// 	}
//
// 	// Output:
// 	// Selected instance: instance-1 with load score: 0.00
// }
//
// // ExamplePerformanceAnalysis shows how to use RuntimeStats for performance analysis.
// func Example_performanceAnalysis() {
// 	type PerformanceReport struct {
// 		MemoryIncrease int64   `json:"memory_increase"`
// 		PeakMemory     int64   `json:"peak_memory"`
// 		GoroutineLeaks int32   `json:"goroutine_leaks"`
// 		CPUUsage       float64 `json:"cpu_usage"`
// 		GCStats        struct {
// 			Count     uint32 `json:"count"`
// 			PauseTime uint64 `json:"pause_time"`
// 		} `json:"gc_stats"`
// 	}
//
// 	ctx := context.Background()
// 	m, _ := monitor.NewRuntimeStats(ctx, nil)
// 	m.Start()
// 	defer m.Stop()
//
// 	// Record baseline
// 	baseLine := m.GetMetrics()
//
// 	// Simulate some work
// 	work := make([]byte, 1024*1024)
// 	for i := 0; i < 100; i++ {
// 		work = append(work, make([]byte, 1024)...)
// 	}
//
// 	// Get current and peak usage
// 	current := m.GetMetrics()
// 	peak := m.GetPeakUsage()
//
// 	report := PerformanceReport{
// 		MemoryIncrease: current.Memory - baseLine.Memory,
// 		PeakMemory:     peak.Memory,
// 		GoroutineLeaks: current.Goroutines - baseLine.Goroutines,
// 		CPUUsage:       current.CPU,
// 	}
// 	report.GCStats.Count = current.GCCount
// 	report.GCStats.PauseTime = current.GCPause
//
// 	out, _ := json.MarshalIndent(report, "", "  ")
// 	fmt.Println(string(out))
//
// 	// Output will vary based on actual runtime usage
// }
//
// type AutoScaler struct{}
//
// func (s *AutoScaler) ScaleUp()   { fmt.Println("Scaling up") }
// func (s *AutoScaler) ScaleDown() { fmt.Println("Scaling down") }
//
// // ExampleAutoscaling demonstrates how to use RuntimeStats for autoscaling decisions.
// func Example_autoscaling() {
//
// 	ctx := context.Background()
// 	m, _ := monitor.NewRuntimeStats(ctx, nil)
// 	m.Start()
// 	defer m.Stop()
//
// 	scaler := &AutoScaler{}
// 	usage := m.GetMetrics()
// 	peak := m.GetPeakUsage()
//
// 	// Make scaling decision based on runtime usage
// 	if usage.Memory > int64(float64(peak.Memory)*0.8) || usage.CPU > 75 {
// 		scaler.ScaleUp()
// 	} else if usage.Memory < int64(float64(peak.Memory)*0.3) && usage.CPU < 30 {
// 		scaler.ScaleDown()
// 	}
//
// 	// Output:
// 	// Scaling down
// }
//
// // ExampleDebugMonitoring shows how to use RuntimeStats for debugging
// // and troubleshooting.
// func Example_debugMonitoring() {
// 	ctx := context.Background()
// 	m, _ := monitor.NewRuntimeStats(ctx, nil)
// 	m.Start()
// 	defer m.Stop()
//
// 	// Get detailed runtime usage
// 	usage := m.GetMetrics()
// 	fmt.Printf("Debug stats:\n"+
// 		"Memory: %d bytes\n"+
// 		"CPU: %.2f%%\n"+
// 		"Goroutines: %d\n"+
// 		"GC Cycles: %d\n"+
// 		"GC Pause: %d ns\n",
// 		usage.Memory,
// 		usage.CPU,
// 		usage.Goroutines,
// 		usage.GCCount,
// 		usage.GCPause)
//
// 	// Output will vary based on actual runtime usage
// }
//
// // Helper function to create monitor
// func mustCreateMonitor(ctx context.Context) *monitor.RuntimeStats {
// 	m, err := monitor.NewRuntimeStats(ctx, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return m
// }

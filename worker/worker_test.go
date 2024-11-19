package worker

//
// import (
// 	"context"
// 	"errors"
// 	"testing"
// 	"time"
// )
//
// // Mock processor
//
// type mockProcessor struct {
// 	processFn func(task any) error
// }
//
// func (m *mockProcessor) Process(task any) error {
// 	return m.processFn(task)
// }
//
// // Worker tests
//
// func TestPool_Submit(t *testing.T) {
// 	p := NewPool(nil, &mockProcessor{
// 		processFn: func(task any) error {
// 			time.Sleep(time.Millisecond * 100)
// 			return nil
// 		},
// 	})
// 	p.Start()
// 	defer p.Stop(context.Background())
//
// 	err := p.Submit("task1")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	err = p.Submit("task2")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	time.Sleep(time.Millisecond * 150)
// 	metrics := p.GetMetrics()
// 	if metrics["completed_tasks"] != 2 {
// 		t.Errorf("expected 2 completed tasks, got %d", metrics["completed_tasks"])
// 	}
// }
//
// func TestPool_SubmitWhenFull(t *testing.T) {
// 	p := NewPool(&Config{
// 		MaxWorkers: 1,
// 		QueueSize:  1,
// 	}, &mockProcessor{
// 		processFn: func(task any) error {
// 			time.Sleep(time.Millisecond * 100)
// 			return nil
// 		},
// 	})
// 	p.Start()
// 	defer p.Stop(context.Background())
//
// 	err := p.Submit("task1")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	err = p.Submit("task2")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	err = p.Submit("task3")
// 	if !errors.Is(err, ErrQueueFull) {
// 		t.Errorf("expected ErrQueueFull, got %v", err)
// 	}
// }
//
// func TestPool_ProcessorError(t *testing.T) {
// 	p := NewPool(nil, &mockProcessor{
// 		processFn: func(task any) error {
// 			return errors.New("processing error")
// 		},
// 	})
// 	p.Start()
// 	defer p.Stop(context.Background())
//
// 	err := p.Submit("task1")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	time.Sleep(time.Millisecond * 50)
// 	metrics := p.GetMetrics()
// 	if metrics["failed_tasks"] != 1 {
// 		t.Errorf("expected 1 failed task, got %d", metrics["failed_tasks"])
// 	}
// }
//
// func TestPool_ProcessorPanic(t *testing.T) {
// 	p := NewPool(nil, &mockProcessor{
// 		processFn: func(task any) error {
// 			panic("processing panic")
// 		},
// 	})
// 	p.Start()
// 	defer p.Stop(context.Background())
//
// 	err := p.Submit("task1")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	time.Sleep(time.Millisecond * 50)
// 	metrics := p.GetMetrics()
// 	if metrics["failed_tasks"] != 1 {
// 		t.Errorf("expected 1 failed task, got %d", metrics["failed_tasks"])
// 	}
// }
//
// func TestPool_ProcessorTimeout(t *testing.T) {
// 	p := NewPool(&Config{
// 		TaskTimeout: time.Millisecond * 50,
// 	}, &mockProcessor{
// 		processFn: func(task any) error {
// 			time.Sleep(time.Millisecond * 100)
// 			return nil
// 		},
// 	})
// 	p.Start()
// 	defer p.Stop(context.Background())
//
// 	err := p.Submit("task1")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	time.Sleep(time.Millisecond * 75)
// 	metrics := p.GetMetrics()
// 	if metrics["failed_tasks"] != 1 {
// 		t.Errorf("expected 1 failed task, got %d", metrics["failed_tasks"])
// 	}
// }
//
// func TestPool_StopWaitsForTaskCompletion(t *testing.T) {
// 	p := NewPool(nil, &mockProcessor{
// 		processFn: func(task any) error {
// 			time.Sleep(time.Millisecond * 100)
// 			return nil
// 		},
// 	})
// 	p.Start()
//
// 	err := p.Submit("task1")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	stopCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond*150)
// 	defer cancel()
// 	p.Stop(stopCtx)
//
// 	metrics := p.GetMetrics()
// 	if metrics["completed_tasks"] != 1 {
// 		t.Errorf("expected 1 completed task, got %d", metrics["completed_tasks"])
// 	}
// }
//
// func TestPool_StopWithTimeout(t *testing.T) {
// 	p := NewPool(nil, &mockProcessor{
// 		processFn: func(task any) error {
// 			time.Sleep(time.Millisecond * 100)
// 			return nil
// 		},
// 	})
// 	p.Start()
//
// 	err := p.Submit("task1")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
//
// 	stopCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
// 	defer cancel()
// 	p.Stop(stopCtx)
//
// 	metrics := p.GetMetrics()
// 	if metrics["completed_tasks"] != 0 {
// 		t.Errorf("expected 0 completed tasks, got %d", metrics["completed_tasks"])
// 	}
// }
//
// // Benchmark tests
//
// func BenchmarkPool_Submit(b *testing.B) {
// 	p := NewPool(nil, &mockProcessor{
// 		processFn: func(task any) error {
// 			time.Sleep(time.Millisecond * 100)
// 			return nil
// 		},
// 	})
// 	p.Start()
// 	defer p.Stop(context.Background())
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		err := p.Submit(i)
// 		if err != nil {
// 			return
// 		}
// 	}
// }
//
// func BenchmarkPool_SubmitWithError(b *testing.B) {
// 	p := NewPool(nil, &mockProcessor{
// 		processFn: func(task any) error {
// 			time.Sleep(time.Millisecond * 100)
// 			return errors.New("processing error")
// 		},
// 	})
// 	p.Start()
// 	defer p.Stop(context.Background())
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		err := p.Submit(i)
// 		if err != nil {
// 			return
// 		}
// 	}
// }
//
// func BenchmarkPool_SubmitWithPanic(b *testing.B) {
// 	p := NewPool(nil, &mockProcessor{
// 		processFn: func(task any) error {
// 			time.Sleep(time.Millisecond * 100)
// 			panic("processing panic")
// 		},
// 	})
// 	p.Start()
// 	defer p.Stop(context.Background())
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		err := p.Submit(i)
// 		if err != nil {
// 			return
// 		}
// 	}
// }

package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type mockProcessor struct {
	processFn func(task any) error
}

func (m *mockProcessor) Process(task any) error {
	return m.processFn(task)
}

func waitForMetric(t *testing.T, p *Pool, name string, want int64) {
	t.Helper()

	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("metric %s did not reach %d, got %d", name, want, p.GetMetrics()[name])
		case <-ticker.C:
			if p.GetMetrics()[name] >= want {
				return
			}
		}
	}
}

func TestNewPoolERejectsInvalidConfig(t *testing.T) {
	_, err := NewPoolE(&Config{MaxWorkers: 0, QueueSize: 1})
	if err == nil {
		t.Fatal("expected invalid configuration error")
	}
}

func TestPoolSubmitCompletesTask(t *testing.T) {
	done := make(chan struct{})
	p := NewPool(nil, &mockProcessor{
		processFn: func(task any) error {
			close(done)
			return nil
		},
	})
	p.Start()
	defer p.Stop(context.Background())

	if err := p.Submit("task"); err != nil {
		t.Fatalf("submit task: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("task was not processed")
	}
	waitForMetric(t, p, "completed_tasks", 1)
}

func TestPoolSubmitWaitBlocksUntilCapacity(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	p := NewPool(&Config{MaxWorkers: 1, QueueSize: 1, TaskTimeout: time.Second}, &mockProcessor{
		processFn: func(task any) error {
			once.Do(func() {
				close(started)
			})
			<-release
			return nil
		},
	})
	p.Start()
	defer p.Stop(context.Background())

	if err := p.Submit("running"); err != nil {
		t.Fatalf("submit running task: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("running task did not start")
	}
	if err := p.Submit("queued"); err != nil {
		t.Fatalf("submit queued task: %v", err)
	}

	submitted := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		submitted <- p.SubmitWait(ctx, "waited")
	}()

	select {
	case err := <-submitted:
		t.Fatalf("SubmitWait returned before capacity was available: %v", err)
	case <-time.After(30 * time.Millisecond):
	}

	close(release)
	if err := <-submitted; err != nil {
		t.Fatalf("SubmitWait after capacity was available: %v", err)
	}
	waitForMetric(t, p, "completed_tasks", 3)
}

func TestPoolProcessorErrorAndPanicAreRecorded(t *testing.T) {
	p := NewPool(nil, &mockProcessor{
		processFn: func(task any) error {
			if task == "panic" {
				panic("processor panic")
			}
			return errors.New("processor error")
		},
	})
	p.Start()
	defer p.Stop(context.Background())

	if err := p.Submit("error"); err != nil {
		t.Fatalf("submit error task: %v", err)
	}
	if err := p.Submit("panic"); err != nil {
		t.Fatalf("submit panic task: %v", err)
	}
	waitForMetric(t, p, "failed_tasks", 2)
}

func TestPoolTaskTimeoutIsRecorded(t *testing.T) {
	p := NewPool(&Config{MaxWorkers: 1, QueueSize: 1, TaskTimeout: 20 * time.Millisecond}, &mockProcessor{
		processFn: func(task any) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	})
	p.Start()
	defer p.Stop(context.Background())

	if err := p.Submit("slow"); err != nil {
		t.Fatalf("submit slow task: %v", err)
	}
	waitForMetric(t, p, "failed_tasks", 1)
}

func TestPoolStopIsIdempotentAndRejectsNewTasks(t *testing.T) {
	p := NewPool(nil)
	p.Start()
	p.Stop(context.Background())
	p.Stop(context.Background())

	if err := p.Submit("after-stop"); !errors.Is(err, ErrPoolStopped) {
		t.Fatalf("expected ErrPoolStopped, got %v", err)
	}
}

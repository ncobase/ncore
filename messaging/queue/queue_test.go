package queue

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func testConfig() *Config {
	return &Config{
		Workers:       1,
		QueueSize:     8,
		TaskTimeout:   time.Second,
		MaxRetries:    0,
		RetryDelay:    time.Millisecond,
		MaxPriority:   10,
		MetricsWindow: 10 * time.Millisecond,
	}
}

func stopQueue(t *testing.T, q *TaskQueue) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := q.Stop(ctx); err != nil {
		t.Fatalf("stop queue: %v", err)
	}
}

func waitForTask(t *testing.T, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("task was not processed")
	}
}

func TestTaskQueueProcessesNormalTask(t *testing.T) {
	done := make(chan struct{})
	q := NewTaskQueue(context.Background(), WithConfig(testConfig()))
	q.Start()
	defer stopQueue(t, q)

	err := q.Push(&QueuedTask{
		ID: "normal",
		Data: func() {
			close(done)
		},
	})
	if err != nil {
		t.Fatalf("push normal task: %v", err)
	}

	waitForTask(t, done)
	if metrics := q.GetMetrics(); metrics["success_count"] != 1 || metrics["dequeue_count"] != 1 {
		t.Fatalf("unexpected metrics: %#v", metrics)
	}
}

func TestTaskQueueAcceptsNilContext(t *testing.T) {
	done := make(chan struct{})
	q := NewTaskQueue(nil, WithConfig(testConfig()))
	q.Start()
	defer stopQueue(t, q)

	err := q.Push(&QueuedTask{
		ID: "nil-context",
		Data: func() {
			close(done)
		},
	})
	if err != nil {
		t.Fatalf("push nil-context task: %v", err)
	}

	waitForTask(t, done)
}

func TestTaskQueuePrioritizesQueuedPriorityTasks(t *testing.T) {
	var (
		mu    sync.Mutex
		order []string
		done  = make(chan struct{})
	)

	record := func(name string) func() {
		return func() {
			mu.Lock()
			order = append(order, name)
			if len(order) == 2 {
				close(done)
			}
			mu.Unlock()
		}
	}

	q := NewTaskQueue(context.Background(), WithConfig(testConfig()))
	defer stopQueue(t, q)

	if err := q.Push(&QueuedTask{ID: "normal", Data: record("normal")}); err != nil {
		t.Fatalf("push normal task: %v", err)
	}
	if err := q.Push(&QueuedTask{ID: "priority", Priority: 10, Data: record("priority")}); err != nil {
		t.Fatalf("push priority task: %v", err)
	}

	q.Start()
	waitForTask(t, done)

	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(order, []string{"priority", "normal"}) {
		t.Fatalf("unexpected execution order: %#v", order)
	}
}

func TestTaskQueueProcessesTimerTask(t *testing.T) {
	done := make(chan struct{})
	q := NewTaskQueue(context.Background(), WithConfig(testConfig()))
	q.Start()
	defer stopQueue(t, q)

	if err := q.Push(&QueuedTask{
		ID:        "timer",
		TriggerAt: time.Now().Add(30 * time.Millisecond),
		Data: func() {
			close(done)
		},
	}); err != nil {
		t.Fatalf("push timer task: %v", err)
	}

	waitForTask(t, done)
}

func TestTaskQueueCancelQueuedTask(t *testing.T) {
	executed := make(chan struct{})
	q := NewTaskQueue(context.Background(), WithConfig(testConfig()))
	defer stopQueue(t, q)

	task := &QueuedTask{
		ID: "queued",
		Data: func() {
			close(executed)
		},
	}
	if err := q.Push(task); err != nil {
		t.Fatalf("push queued task: %v", err)
	}
	if !q.Cancel(task.ID) {
		t.Fatal("expected queued task cancellation to succeed")
	}

	q.Start()

	select {
	case <-executed:
		t.Fatal("cancelled task was executed")
	case <-time.After(100 * time.Millisecond):
	}
	if task.Status != TaskStatusCanceled {
		t.Fatalf("expected canceled status, got %s", task.Status)
	}
}

func TestTaskQueueRejectsDuplicateAndStoppedTasks(t *testing.T) {
	q := NewTaskQueue(context.Background(), WithConfig(testConfig()))

	if err := q.Push(&QueuedTask{ID: "duplicate", Data: func() {}}); err != nil {
		t.Fatalf("push first duplicate task: %v", err)
	}
	if err := q.Push(&QueuedTask{ID: "duplicate", Data: func() {}}); !errors.Is(err, ErrDuplicateTask) {
		t.Fatalf("expected ErrDuplicateTask, got %v", err)
	}

	stopQueue(t, q)
	if err := q.Push(&QueuedTask{ID: "after-stop", Data: func() {}}); !errors.Is(err, ErrQueueStopped) {
		t.Fatalf("expected ErrQueueStopped, got %v", err)
	}
}

func TestTaskQueueStopCancelsPendingTasks(t *testing.T) {
	q := NewTaskQueue(context.Background(), WithConfig(testConfig()))
	task := &QueuedTask{ID: "pending", Priority: 5, Data: func() {}}
	if err := q.Push(task); err != nil {
		t.Fatalf("push pending task: %v", err)
	}

	stopQueue(t, q)

	if task.Status != TaskStatusCanceled {
		t.Fatalf("expected pending task to be canceled, got %s", task.Status)
	}
	if q.GetMetrics()["queue_length"] != 0 {
		t.Fatalf("expected queue length to be reset, got %#v", q.GetMetrics())
	}
}

# Example 05: Background Jobs & Worker Pools

Demonstrates async task processing with NCore's worker pools, job queuing, and status tracking patterns used in production systems.

## Features

- **Worker Pool**: Concurrent job processing with `concurrency/worker`
- **Job Queue**: FIFO job scheduling
- **Status Tracking**: Monitor job progress and results
- **Graceful Shutdown**: Clean worker termination
- **Error Handling**: Retry logic and error reporting
- **Job Types**: Multiple job handlers

## Architecture

```text
┌─────────────┐      Submit      ┌──────────────┐
│   HTTP API  │─────────────────►│  Job Queue   │
└─────────────┘                  └──────┬───────┘
                                        │
                                        │ Dequeue
                                        ▼
                              ┌──────────────────┐
                              │   Worker Pool    │
                              │ ┌──────┬──────┐ │
                              │ │ W1   │  W2  │ │
                              │ ├──────┼──────┤ │
                              │ │ W3   │  W4  │ │
                              │ └──────┴──────┘ │
                              └──────────────────┘
                                        │
                                        │ Update
                                        ▼
                              ┌──────────────────┐
                              │  Status Tracker  │
                              └──────────────────┘
```

## Features Demonstrated

### 1. Worker Pool Configuration

```go
pool, cleanup, err := worker.New(&worker.Config{
    MaxWorkers: 10,
    QueueSize:  100,
})
defer cleanup()
```

### 2. Job Submission

```go
job := &Job{
    ID:     uuid.New().String(),
    Type:   "email",
    Payload: emailData,
    Status: "pending",
}

pool.Submit(context.Background(), func(ctx context.Context) error {
    return processEmail(ctx, job)
})
```

### 3. Status Tracking

```go
type JobTracker struct {
    jobs map[string]*JobStatus
    mu   sync.RWMutex
}

func (t *JobTracker) UpdateStatus(jobID string, status string, progress int) {
    t.mu.Lock()
    defer t.mu.Unlock()

    if job, exists := t.jobs[jobID]; exists {
        job.Status = status
        job.Progress = progress
        job.UpdatedAt = time.Now()
    }
}
```

## Project Structure

```text
05-background-jobs/
├── job/
│   ├── data/
│   │   ├── data.go          # SQLite connection
│   │   └── repository/
│   │       └── job.go       # SQLite job store
│   ├── handler/
│   │   └── job.go           # HTTP handlers
│   ├── structs/
│   │   └── job.go           # Job models
│   ├── handlers.go          # Built-in job handlers
│   └── manager.go           # Worker pool orchestration
├── main.go
├── config.yaml
└── README.md
```

## Job Types

### 1. Email Job

```bash
curl -X POST http://localhost:8080/jobs \
  -d '{
    "type": "email",
    "payload": {
      "to": "user@example.com",
      "subject": "Hello",
      "body": "Test email"
    }
  }'
```

### 2. Data Export Job

```bash
curl -X POST http://localhost:8080/jobs \
  -d '{
    "type": "export",
    "payload": {
      "format": "csv",
      "filters": {"date": "2024-01"}
    }
  }'
```

### 3. Cleanup Job

```bash
curl -X POST http://localhost:8080/jobs \
  -d '{
    "type": "cleanup",
    "payload": {
      "older_than": "30d"
    }
  }'
```

## Checking Job Status

```bash
curl http://localhost:8080/jobs/abc-123-def

# Response:
{
  "id": "abc-123-def",
  "type": "export",
  "status": "processing",
  "progress": 45,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:05:00Z"
}
```

## Worker Pool Patterns

### 1. Dynamic Scaling

```go
// Start with minimum workers
pool.SetWorkers(5)

// Scale up during peak hours
if isHighLoad() {
    pool.SetWorkers(20)
}
```

### 2. Priority Queue

```go
type PriorityJob struct {
    Job      *Job
    Priority int
}

// High priority jobs processed first
queue.Push(&PriorityJob{job, 1}) // High
queue.Push(&PriorityJob{job, 5}) // Low
```

### 3. Retry Logic

```go
func processWithRetry(ctx context.Context, job *Job) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        if err := process(ctx, job); err != nil {
            if i == maxRetries-1 {
                return err
            }
            time.Sleep(time.Second * time.Duration(i+1))
            continue
        }
        return nil
    }
    return nil
}
```

```yaml
data:
  database:
    master:
      driver: sqlite3
      source: "file:jobs.db?cache=shared&_fk=1"

worker:
  max_workers: 10
  queue_size: 100
  job_timeout: 300 # 5 minutes

jobs:
  email:
    retry_count: 3
    timeout: 60
  export:
    retry_count: 1
    timeout: 600 # 10 minutes
```

## Monitoring

```bash
# Get worker pool stats
curl http://localhost:8080/workers/stats

# Response:
{
  "active_workers": 7,
  "queued_jobs": 12,
  "completed_jobs": 156,
  "failed_jobs": 3
}
```

## Testing

```bash
# Unit tests
go test ./job/...
```

## Use Cases

- Email sending queues
- Data export/import
- Image processing
- Report generation
- Database cleanup
- Batch operations

## Next Steps

- Add [event notifications](../06-event-driven) on job completion
- Integrate with [Redis queue](../08-full-application)
- Add [authentication](../07-authentication) for job management

## License

This example is part of the NCore project.

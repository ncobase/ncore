// Package job coordinates background job execution and storage.
package job

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/concurrency/worker"
	jobRepo "github.com/ncobase/ncore/examples/05-background-jobs/job/data/repository"
	"github.com/ncobase/ncore/examples/05-background-jobs/job/structs"
	"github.com/ncobase/ncore/logging/logger"
)

type Manager struct {
	pool     *worker.Pool
	repo     jobRepo.JobRepository
	logger   *logger.Logger
	handlers map[string]JobHandler
}

type JobHandler func(ctx context.Context, job *structs.Job, updateProgress func(int)) error

func NewManager(cfg *worker.Config, repo jobRepo.JobRepository, logger *logger.Logger) (*Manager, func(), error) {
	processor := &jobProcessor{}
	pool := worker.NewPool(cfg, processor)
	pool.Start()

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		pool.Stop(ctx)
	}

	return &Manager{
		pool:     pool,
		repo:     repo,
		logger:   logger,
		handlers: make(map[string]JobHandler),
	}, cleanup, nil
}

type jobProcessor struct{}

func (p *jobProcessor) Process(task any) error {
	if fn, ok := task.(func() error); ok {
		return fn()
	}
	return fmt.Errorf("invalid task type")
}

func (m *Manager) RegisterHandler(jobType string, handler JobHandler) {
	m.handlers[jobType] = handler
}

func (m *Manager) Submit(ctx context.Context, jobType string, payload map[string]any) (*structs.Job, error) {
	job := &structs.Job{
		ID:        uuid.New().String(),
		Type:      jobType,
		Payload:   payload,
		Status:    structs.StatusPending,
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	err := m.pool.Submit(func() error {
		return m.executeJob(ctx, job)
	})
	if err != nil {
		if updateErr := m.updateJobStatus(ctx, job, structs.StatusFailed, err.Error()); updateErr != nil {
			m.logger.Error(ctx, "Failed to update job status", "error", updateErr)
		}
		return nil, err
	}

	m.logger.Info(ctx, "Job submitted", "job_id", job.ID, "type", jobType)
	return job, nil
}

func (m *Manager) executeJob(ctx context.Context, job *structs.Job) error {
	m.logger.Info(ctx, "Executing job", "job_id", job.ID, "type", job.Type)

	now := time.Now()
	job.StartedAt = &now
	if err := m.updateJobStatus(ctx, job, structs.StatusRunning, ""); err != nil {
		return err
	}

	handler, ok := m.handlers[job.Type]
	if !ok {
		err := fmt.Errorf("no handler for job type: %s", job.Type)
		_ = m.updateJobStatus(ctx, job, structs.StatusFailed, err.Error())
		return err
	}

	updateProgress := func(progress int) {
		job.Progress = progress
		job.UpdatedAt = time.Now()
		if err := m.repo.Update(ctx, job); err != nil {
			m.logger.Error(ctx, "Failed to update job progress", "error", err, "job_id", job.ID)
		}
	}

	err := handler(ctx, job, updateProgress)

	endTime := time.Now()
	job.EndedAt = &endTime

	if err != nil {
		_ = m.updateJobStatus(ctx, job, structs.StatusFailed, err.Error())
		m.logger.Error(ctx, "Job failed", "job_id", job.ID, "error", err)
		return err
	}

	if err := m.updateJobStatus(ctx, job, structs.StatusCompleted, ""); err != nil {
		return err
	}
	m.logger.Info(ctx, "Job completed", "job_id", job.ID)
	return nil
}

func (m *Manager) updateJobStatus(ctx context.Context, job *structs.Job, status structs.JobStatus, errMsg string) error {
	job.Status = status
	job.UpdatedAt = time.Now()
	if errMsg != "" {
		job.Error = errMsg
	}
	if status == structs.StatusCompleted {
		job.Progress = 100
	}
	return m.repo.Update(ctx, job)
}

func (m *Manager) GetJob(jobID string) (*structs.Job, error) {
	return m.repo.FindByID(context.Background(), jobID)
}

func (m *Manager) ListJobs() []*structs.Job {
	jobs, err := m.repo.List(context.Background())
	if err != nil {
		m.logger.Error(context.Background(), "Failed to list jobs", "error", err)
		return []*structs.Job{}
	}
	return jobs
}

func (m *Manager) GetStats() map[string]any {
	stats, err := m.repo.Stats(context.Background())
	if err != nil {
		m.logger.Error(context.Background(), "Failed to fetch job stats", "error", err)
		return map[string]any{}
	}

	return map[string]any{
		"total":     stats[structs.StatusPending] + stats[structs.StatusRunning] + stats[structs.StatusCompleted] + stats[structs.StatusFailed],
		"pending":   stats[structs.StatusPending],
		"running":   stats[structs.StatusRunning],
		"completed": stats[structs.StatusCompleted],
		"failed":    stats[structs.StatusFailed],
	}
}

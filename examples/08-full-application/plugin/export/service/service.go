// Package service contains export job business logic.
package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ncobase/ncore/concurrency/worker"
	"github.com/ncobase/ncore/examples/08-full-application/biz/task"
	taskrepo "github.com/ncobase/ncore/examples/08-full-application/biz/task/data/repository"
	"github.com/ncobase/ncore/examples/08-full-application/internal/event"
	exportrepo "github.com/ncobase/ncore/examples/08-full-application/plugin/export/data/repository"
	"github.com/ncobase/ncore/examples/08-full-application/plugin/export/structs"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

type Service struct {
	jobRepo  exportrepo.JobRepository
	taskRepo taskrepo.TaskRepository
	pool     *worker.Pool
	bus      *event.Bus
	logger   *logger.Logger
	jobs     map[string]*structs.Job

	mu *sync.Map
}

func NewService(logger *logger.Logger, bus *event.Bus, pool *worker.Pool) *Service {
	return &Service{
		bus:    bus,
		pool:   pool,
		logger: logger,
		jobs:   make(map[string]*structs.Job),
		mu:     &sync.Map{},
	}
}

func (s *Service) SetRepositories(jobRepo exportrepo.JobRepository, taskRepo taskrepo.TaskRepository) {
	s.jobRepo = jobRepo
	s.taskRepo = taskRepo
}

func (s *Service) CreateExport(ctx context.Context, workspaceID, userID string, req *structs.CreateExportRequest) (*structs.Job, error) {
	job := &structs.Job{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Type:        req.Type,
		Format:      req.Format,
		Status:      structs.JobStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		s.logger.Error(ctx, "Failed to create export job", "error", err)
		return nil, err
	}

	s.mu.Store(job.ID, job)

	jobID := job.ID
	if err := s.pool.Submit(func() error {
		return s.executeExport(ctx, jobID)
	}); err != nil {
		s.logger.Error(ctx, "Failed to submit export job", "error", err, "job_id", jobID)
		return nil, err
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          event.EventTypeExportRequested,
		AggregateID:   job.ID,
		AggregateName: "export_job",
		WorkspaceID:   workspaceID,
		UserID:        userID,
		Payload: map[string]any{
			"job_id": job.ID,
			"type":   req.Type,
			"format": req.Format,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish export requested event", "error", err)
	}

	s.logger.Info(ctx, "Export job created", "job_id", job.ID, "workspace_id", workspaceID, "type", req.Type)
	return job, nil
}

func (s *Service) GetJob(ctx context.Context, jobID string) (*structs.Job, error) {
	if job, ok := s.mu.Load(jobID); ok {
		return job.(*structs.Job), nil
	}

	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get export job", "error", err, "job_id", jobID)
		return nil, err
	}

	return job, nil
}

func (s *Service) ListJobs(ctx context.Context, workspaceID string, limit, offset int) ([]*structs.Job, error) {
	jobs, err := s.jobRepo.FindByWorkspace(ctx, workspaceID, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to list export jobs", "error", err, "workspace_id", workspaceID)
		return nil, err
	}
	return jobs, nil
}

func (s *Service) executeExport(ctx context.Context, jobID string) error {
	jobVal, ok := s.mu.Load(jobID)
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job := jobVal.(*structs.Job)

	job.Status = structs.JobStatusRunning
	job.UpdatedAt = time.Now()
	s.mu.Store(jobID, job)

	var filePath string
	var err error

	switch job.Type {
	case "tasks":
		filePath, err = s.exportTasks(ctx, job.WorkspaceID, job.Format)
	case "comments":
		filePath, err = s.exportComments(ctx, job.WorkspaceID, job.Format)
	default:
		return fmt.Errorf("unsupported export type: %s", job.Type)
	}

	now := time.Now()
	job.UpdatedAt = now
	job.CompletedAt = &now

	if err != nil {
		job.Status = structs.JobStatusFailed
		job.Error = err.Error()
		s.logger.Error(ctx, "Export job failed", "error", err, "job_id", jobID)
	} else {
		job.Status = structs.JobStatusCompleted
		job.FilePath = filePath
		s.logger.Info(ctx, "Export job completed", "job_id", jobID, "file_path", filePath)
	}

	s.mu.Store(jobID, job)

	eventType := event.EventTypeExportCompleted
	if err != nil {
		eventType = event.EventTypeExportFailed
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          eventType,
		AggregateID:   job.ID,
		AggregateName: "export_job",
		WorkspaceID:   job.WorkspaceID,
		UserID:        job.UserID,
		Payload: map[string]any{
			"job_id":    job.ID,
			"status":    job.Status,
			"file_path": job.FilePath,
			"error":     job.Error,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish export completed event", "error", err)
	}

	return err
}

func (s *Service) exportTasks(ctx context.Context, workspaceID, format string) (string, error) {
	tasks, err := s.taskRepo.FindByWorkspace(ctx, workspaceID, 1000, 0)
	if err != nil {
		return "", fmt.Errorf("failed to fetch tasks: %w", err)
	}

	filePath := fmt.Sprintf("/tmp/export_%s_tasks_%s.%s", workspaceID, time.Now().Format("20060102_150405"), format)

	switch format {
	case "csv":
		if err := s.writeTasksCSV(tasks, filePath); err != nil {
			return "", err
		}
	case "json":
		if err := s.writeTasksJSON(tasks, filePath); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

func (s *Service) exportComments(ctx context.Context, workspaceID, format string) (string, error) {
	_ = ctx
	filePath := fmt.Sprintf("/tmp/export_%s_comments_%s.%s", workspaceID, time.Now().Format("20060102_150405"), format)

	switch format {
	case "csv":
	case "json":
	}

	return filePath, nil
}

func (s *Service) writeTasksCSV(tasks []*task.Task, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"ID", "Title", "Description", "Status", "Priority", "Assigned To", "Created By", "Created At", "Updated At"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, task := range tasks {
		row := []string{
			task.ID,
			task.Title,
			task.Description,
			task.Status,
			task.Priority,
			task.AssignedTo,
			task.CreatedBy,
			task.CreatedAt.Format(time.RFC3339),
			task.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) writeTasksJSON(tasks []*task.Task, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(tasks); err != nil {
		return err
	}

	return nil
}

func (s *Service) HandleCreate(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	userID, _ := c.Get("user_id")

	var req structs.CreateExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	job, err := s.CreateExport(c.Request.Context(), workspaceID, userID.(string), &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to create export job"))
		return
	}

	resp.WithStatusCode(c.Writer, 201, job)
}

func (s *Service) HandleGetByID(c *gin.Context) {
	jobID := c.Param("job_id")

	job, err := s.GetJob(c.Request.Context(), jobID)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("export job not found"))
		return
	}

	resp.Success(c.Writer, job)
}

func (s *Service) HandleList(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	limit := 20
	offset := 0

	jobs, err := s.ListJobs(c.Request.Context(), workspaceID, limit, offset)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list export jobs"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"jobs":   jobs,
		"limit":  limit,
		"offset": offset,
		"count":  len(jobs),
	})
}

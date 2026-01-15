// Package handler provides HTTP endpoints for job management.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/05-background-jobs/job"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// JobHandler handles job HTTP requests.
type JobHandler struct {
	manager *job.Manager
	logger  *logger.Logger
}

// NewJobHandler creates a new job handler.
func NewJobHandler(mgr *job.Manager, logger *logger.Logger) *JobHandler {
	return &JobHandler{
		manager: mgr,
		logger:  logger,
	}
}

// CreateJob handles job creation.
func (h *JobHandler) CreateJob(c *gin.Context) {
	var req struct {
		Type    string         `json:"type" binding:"required"`
		Payload map[string]any `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	job, err := h.manager.Submit(c.Request.Context(), req.Type, req.Payload)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to submit job"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusAccepted, job)
}

// GetJob retrieves a job by ID.
func (h *JobHandler) GetJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.manager.GetJob(id)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("job not found"))
		return
	}

	resp.Success(c.Writer, job)
}

// ListJobs lists all jobs.
func (h *JobHandler) ListJobs(c *gin.Context) {
	jobs := h.manager.ListJobs()
	resp.Success(c.Writer, jobs)
}

// GetStats returns job statistics.
func (h *JobHandler) GetStats(c *gin.Context) {
	stats := h.manager.GetStats()
	resp.Success(c.Writer, stats)
}

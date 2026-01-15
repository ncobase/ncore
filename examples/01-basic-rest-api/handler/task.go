package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// TaskHandler handles HTTP requests for tasks.
type TaskHandler struct {
	svc    *service.TaskService
	logger *logger.Logger
}

// NewTaskHandler creates a new task handler.
func NewTaskHandler(svc *service.TaskService, logger *logger.Logger) *TaskHandler {
	return &TaskHandler{
		svc:    svc,
		logger: logger,
	}
}

// Create handles task creation.
// @Summary Create a new task
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body service.CreateTaskRequest true "Create task request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "invalid request", "error", err)
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	task, err := h.svc.CreateTask(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error(c.Request.Context(), "failed to create task", "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to create task"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusCreated, task)
}

// Get handles task retrieval.
// @Summary Get a task by ID
// @Tags tasks
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tasks/{task_id} [get]
func (h *TaskHandler) Get(c *gin.Context) {
	idStr := c.Param("task_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("invalid task ID"))
		return
	}

	task, err := h.svc.GetTask(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "task not found" {
			resp.Fail(c.Writer, resp.NotFound("task not found"))
			return
		}
		h.logger.Error(c.Request.Context(), "failed to get task", "id", id, "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to get task"))
		return
	}

	resp.Success(c.Writer, task)
}

// List handles task listing with pagination.
// @Summary List tasks
// @Tags tasks
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	tasks, total, err := h.svc.ListTasks(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error(c.Request.Context(), "failed to list tasks", "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to list tasks"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"data": tasks,
		"pagination": map[string]any{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// Update handles task updates.
// @Summary Update a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body service.UpdateTaskRequest true "Update task request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tasks/{task_id} [put]
func (h *TaskHandler) Update(c *gin.Context) {
	idStr := c.Param("task_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("invalid task ID"))
		return
	}

	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "invalid request", "error", err)
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	task, err := h.svc.UpdateTask(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "task not found" {
			resp.Fail(c.Writer, resp.NotFound("task not found"))
			return
		}
		h.logger.Error(c.Request.Context(), "failed to update task", "id", id, "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to update task"))
		return
	}

	resp.Success(c.Writer, task)
}

// Delete handles task deletion.
// @Summary Delete a task
// @Tags tasks
// @Produce json
// @Param id path int true "Task ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tasks/{task_id} [delete]
func (h *TaskHandler) Delete(c *gin.Context) {
	idStr := c.Param("task_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("invalid task ID"))
		return
	}

	if err := h.svc.DeleteTask(c.Request.Context(), id); err != nil {
		if err.Error() == "task not found" {
			resp.Fail(c.Writer, resp.NotFound("task not found"))
			return
		}
		h.logger.Error(c.Request.Context(), "failed to delete task", "id", id, "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to delete task"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusNoContent)
}

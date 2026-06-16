package tasks_transport_http

import (
	"fmt"
	"net/http"

	"github.com/Lyzix0/todoapp/internal/core/domain"
	core_logger "github.com/Lyzix0/todoapp/internal/core/logger"
	core_http_request "github.com/Lyzix0/todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Lyzix0/todoapp/internal/core/transport/http/response"
	core_http_types "github.com/Lyzix0/todoapp/internal/core/transport/http/types"
)

type PatchTaskRequest struct {
	Title       core_http_types.Nullable[string] `json:"title"`
	Description core_http_types.Nullable[string] `json:"description"`
	Completed   core_http_types.Nullable[bool]   `json:"completed"`
}

func (r *PatchTaskRequest) Validate() error {
	if r.Title.Set {
		if r.Title.Value == nil {
			return fmt.Errorf("`title` can't be null")
		}

		titleLen := len([]rune(*r.Title.Value))
		if titleLen < 1 || titleLen > 100 {
			return fmt.Errorf("'title' must be between 1 and 100 symbols")
		}
	}

	if r.Description.Set {
		if r.Description.Value != nil {
			descriptionLen := len([]rune(*r.Description.Value))
			if descriptionLen < 1 || descriptionLen > 1000 {
				return fmt.Errorf("'description' must be between 1 and 1000 symbols")
			}
		}
	}

	if r.Completed.Set {
		if r.Completed.Value == nil {
			return fmt.Errorf("'Completed' can't be null")
		}
	}

	return nil
}

type PatchTaskResponse TaskDTOResponse

// PatchTask godoc
// @Summary Edit task
// @Description Edit task information in database
// @Description ### Three-state logic:
// @Description 1. Field wasn't in request: field is ignored
// @Description 2. Field was transferred: database changes it
// @Description 3. Field = null: database clears it
// @Description 4. `title` and `complete` can't be null
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body PatchTaskRequest true "PatchTask request body"
// @Success 200 {object} PatchTaskResponse "Successfully edit task"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 409 {object} core_http_response.ErrorResponse "Conflict"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /tasks/{id} [patch]
func (h *TasksHTTPHandler) PatchTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	taskID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get taskID path value",
		)

		return
	}

	var request PatchTaskRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)

		return
	}

	taskPatch := taskPatchFromRequest(request)
	taskDomain, err := h.tasksService.PatchTask(ctx, taskID, taskPatch)

	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to patch task",
		)

		return
	}

	response := PatchTaskResponse(taskDTOFromDomain(taskDomain))
	responseHandler.JSONResponse(response, http.StatusOK)
}

func taskPatchFromRequest(request PatchTaskRequest) domain.TaskPatch {
	return domain.NewTaskPatch(
		request.Title.ToDomain(),
		request.Description.ToDomain(),
		request.Completed.ToDomain(),
	)
}

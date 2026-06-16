package stats_transport_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lyzix0/todoapp/internal/core/domain"
	core_logger "github.com/Lyzix0/todoapp/internal/core/logger"
	core_http_request "github.com/Lyzix0/todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Lyzix0/todoapp/internal/core/transport/http/response"
)

type GetStatsResponse struct {
	TasksCreated               int      `json:"tasks_created"`
	TasksCompleted             int      `json:"tasks_completed"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate"`
	TasksAverageCompletionTime *string  `json:"tasks_averate_completion_time"`
}

// GetStats godoc
// @Summary get stats
// @Description get tasks statistics for all tasks or by UserID
// @Tags stats
// @Produce json
// @Param user_id query int false "filter stats by userID"
// @Param from query string false "starting interval"
// @Param to query string false "end interval"
// @Success 200 {object} GetStatsResponse "Successfull get statistics"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /stats [get]
func (h *StatsHTTPHandler) GetStats(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	queryParams, err := getUserIDFromToQueryParams(r)
	userID, from, to := queryParams.userID, queryParams.from, queryParams.to

	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get userID/from/to query params",
		)

		return
	}

	stats, err := h.statsService.GetStats(ctx, userID, from, to)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get stats",
		)

		return
	}

	response := toDTOFromDomain(stats)
	responseHandler.JSONResponse(response, http.StatusOK)
}

type queryParams struct {
	userID *int
	from   *time.Time
	to     *time.Time
}

func toDTOFromDomain(stats domain.Stats) GetStatsResponse {
	var avgTime *string
	if stats.TasksAverageCompletionTime != nil {
		duration := stats.TasksAverageCompletionTime.String()
		avgTime = &duration
	}

	return GetStatsResponse{
		TasksCreated:               stats.TasksCreated,
		TasksCompleted:             stats.TasksCompleted,
		TasksCompletedRate:         stats.TasksCompletedRate,
		TasksAverageCompletionTime: avgTime,
	}
}

func getUserIDFromToQueryParams(r *http.Request) (queryParams, error) {
	const (
		userIDQueryParamKey = "user_id"
		fromQueryParamKey   = "from"
		toQueryParamKey     = "to"
	)

	userID, err := core_http_request.GetIntQueryParam(r, userIDQueryParamKey)
	if err != nil {
		return queryParams{}, fmt.Errorf("get 'user_id' query param: %w", err)
	}

	from, err := core_http_request.GetDateQueryParam(r, fromQueryParamKey)
	if err != nil {
		return queryParams{}, fmt.Errorf("get 'from' query param: %w", err)
	}

	to, err := core_http_request.GetDateQueryParam(r, toQueryParamKey)
	if err != nil {
		return queryParams{}, fmt.Errorf("get 'to' query param: %w", err)
	}

	return queryParams{userID, from, to}, nil
}

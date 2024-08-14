package api

import (
	db "eduApp/db/sqlc"
	"eduApp/token"
	"eduApp/worker"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

// CreateRequestRequest defines the request body structure for creating a request to delete a course
type CreateRequestRequest struct {
	CourseID int64 `json:"course_id"`
	Pending  bool  `json:"pending"`
}

// @Summary Create a new request
// @Description Create a new request
// @ID create-request
// @Accept  json
// @Produce  json
// @Param request body CreateRequestRequest true "Create Request "
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /request/create [post]
// CreateRequest creates a new request to delete a course
func (server *Server) CreateRequest(ctx *gin.Context) {
	var req CreateRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateRequestParams{
		CourseID: req.CourseID,
		Pending:  true,
	}

	request, err := server.store.CreateRequest(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, request)
}

// UpdateRequestRequest defines the request body structure for updating a request
type UpdateRequestRequest struct {
	Confirm   bool  `json:"confirm"`
	Pending   bool  `json:"pending"`
	CourseID  int64 `json:"course_id"`
	RequestID int64 `json:"request_id"`
}

// @Summary Update a request
// @Description Update a request
// @ID update-request
// @Accept  json
// @Produce  json
// @Param request_id path int true "Request ID"
// @Param request body UpdateRequestRequest true "Update Request Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /requests/edit [put]
// UpdateRequest updates a request
func (server *Server) UpdateRequest(ctx *gin.Context) {
	var req UpdateRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enqueue the update task
	taskPayload := &worker.PayloadUpdateRequest{
		RequestID: req.RequestID,
	}

	// Ensure authorization payload is available in the context
	authPayload, exists := ctx.Get(authorizationPayloadKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization payload"})
		return
	}
	taskPayload.GetPayload = authPayload.(*token.Payload)

	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}
	// Enqueue the update task
	err := server.taskDistributor.DistributeTaskUpdateRequest(ctx, taskPayload, opts...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Return immediately after enqueuing the task
	ctx.JSON(http.StatusOK, gin.H{"message": "Request update task enqueued successfully"})
}

package api

import (
	"database/sql"
	db "eduApp/db/sqlc"
	"eduApp/token"
	"eduApp/worker"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetSubscriptionRequest defines the request body structure for getting a subscription
type GetSubscriptionRequest struct {
	UserID int64 `form:"user_id"`
}

// @Summary Get a subscription By User
// @Description Retrieves a subscription by user ID
// @ID GetSubscription
// @Produce json
// @Param user_id path int true "UserID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /subscription/get [get]
// GetSubscription retrieves a Subscription by user
func (server *Server) GetSubscription(ctx *gin.Context) {
	var req GetSubscriptionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	getSubscription, err := server.store.GetSubscription(ctx, req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, getSubscription)
}

// ListSubscriptionsByUserRequest defines the request body structure for listing subscriptions by user
type ListSubscriptionsByUserRequest struct {
	CourseID int64 `form:"course_id"`
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=100"`
}

// @Summary List subscriptions
// @Description List subscriptions for a given Course ID
// @ID listSubscriptions
// @Produce json
// @Param course_id path int true "Course ID"
// @Param limit query int true "Limit" minimum(1) maximum(100)
// @Param offset query int true "Offset" minimum(0)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /subscriptions/user [get]
// ListSubscriptionsByUser lists subscriptions by user
func (server *Server) ListSubscriptionsByUser(ctx *gin.Context) {
	var req ListSubscriptionsByUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListSubscriptionsByUserParams{
		CourseID: req.CourseID,
		Limit:    req.PageSize,
		Offset:   (req.PageID - 1) * req.PageSize,
	}

	subscriptions, err := server.store.ListSubscriptionsByUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, subscriptions)
}

// ListSubscriptionsByCourseRequest defines the request body structure for listing subscriptions by user
type ListSubscriptionsByCourseRequest struct {
	UserID   int64 `form:"user_id"`
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=100"`
}

// @Summary List subscriptions
// @Description List subscriptions for a given user ID
// @ID listSubscriptions
// @Produce json
// @Param course_id path int true "User ID"
// @Param limit query int true "Limit" minimum(1) maximum(100)
// @Param offset query int true "Offset" minimum(0)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /subscriptions/user [get]
// ListSubscriptionsByCourse lists subscriptions by course
func (server *Server) ListSubscriptionsByCourse(ctx *gin.Context) {
	var req ListSubscriptionsByCourseRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListSubscriptionsByCourseParams{
		UserID: req.UserID,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	subscriptions, err := server.store.ListSubscriptionsByCourse(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, subscriptions)
}

// UpdateSubscriptionsRequest defines the request body structure for updating a subscription
type UpdateSubscriptionsRequest struct {
	Active   bool  `json:"active"`
	Pending  bool  `json:"pending"`
	UserID   int64 `json:"user_id"`
	CourseID int64 `json:"course_id"`
}

// @Summary Update a Subscription
// @Description Update a subscription
// @ID update-subscription
// @Accept  json
// @Produce  json
// @Param course_id path int true "Course ID"
// @Param request body UpdateSubscriptionsRequest true "Update Subscription Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /subscription/edit [put]
// UpdateSubscriptions updates a subscription
func (server *Server) UpdateSubscriptions(ctx *gin.Context) {
	var req UpdateSubscriptionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.UpdateSubscriptionsParams{
		UserID: req.UserID,
		Active: pgtype.Bool{
			Bool:  req.Active,
			Valid: true,
		},
		Pending: pgtype.Bool{
			Bool:  req.Pending,
			Valid: true,
		},
		CourseID: req.CourseID,
	}

	material, err := server.store.UpdateSubscriptions(ctx, db.UpdateSubscriptionsParams(arg))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, material)
}

type createSubscriptionRequest struct {
	UserID   int64 `json:"user_id"`
	CourseID int64 `json:"course_id"`
}

// @Summary Create a new Subscription
// @Description Creates a new Subscription by depend on course
// @Accept json
// @Produce json
// @Param request body CreateSubscriptionRequest true "user_id and course_id"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /subscription [post]
// CreateSubscription creates a new course subscription

func (server Server) CreateSubscription(ctx *gin.Context) {
	var req createSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if req.UserID != authPayload.UserID {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.CreateSubscriptionTxParams{
		CreateSubscriptionParams: db.CreateSubscriptionParams{
			UserID:   req.UserID,
			CourseID: req.CourseID,
			Active:   false,
			Pending:  true,
		},
		AfterCreate: func(subscription db.Subscription) error {
			// Use Redis for task distribution
			taskPayload := &worker.PayloadCreateSubscription{
				UserID:   subscription.UserID,
				CourseID: subscription.CourseID,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskCreateSubscription(ctx, taskPayload, opts...)
		},
	}

	txResult, err := server.store.CreateSubscriptionTx(ctx, arg)

	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolations {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, txResult)
}

// GetUserCountForCertianCourseRequest contains input parameter for get total subscribed user count for certian course only
type GetUserCountForCertianCourseRequest struct {
	Active   bool  `json:"active"`
	CourseID int64 `form:"course_id"`
}

// @Summary Get User Count for certian course
// @Description Get Subscribed user count for certian course
// @Accept json
// @Produce json
// @Param request body GetUserCountForCertianCourseRequest true "Get User Count For Certain Course"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /count/course/subscription [Get]
// GetUserCountForCertianCourse is a funtion api call to get all subscribed users for certian curse only
func (server *Server) GetUserCountForCertianCourse(ctx *gin.Context) {
	var req GetUserCountForCertianCourseRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.GetUserCountForCertianCourseParams{
		Active:   true,
		CourseID: req.CourseID,
	}

	count, err := server.store.GetUserCountForCertianCourse(ctx, db.GetUserCountForCertianCourseParams(arg))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Get registered user count for this course,  Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, count)
}

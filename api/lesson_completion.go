package api

import (
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

// UpdateLessonCompletionRequest defines the request body structure for listing courses
type UpdateLessonCompletionRequest struct {
	CourseID   int64 `json:"course_id"`
	UserID     int64 `json:"user_id"`
	MaterialID int64 `json:"material_id"`
}

// @Summary Update a UpdateLessonCompletion
// @Description Updates a UpdateLessonCompletion with provided details
// @Accept json
// @Produce json
// @Param request body UpdateLessonCompletionRequest true "Updated UpdateLessonCompletion details"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /lessonCompletion/edit [put]
// UpdateCourse updates the selected course
func (server *Server) UpdateLessonCompletion(ctx *gin.Context) {
	var req UpdateLessonCompletionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.UserID != req.UserID {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.UpdateLessonCompletionParams{
		CourseID:   req.CourseID,
		UserID:     req.UserID,
		MaterialID: req.MaterialID,
		Completed: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}

	LessonCompletion, err := server.store.UpdateLessonCompletion(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, LessonCompletion)
}

// GetTotalMaterialInCourseRequest defines the request body structure for getting materials
type GetTotalLessonsCompleteInCourseRequest struct {
	CourseID int64 `form:"course_id"`
}

// @Summary Get materials for a course
// @Description Get materials for a course
// @ID get-materials
// @Accept  json
// @Produce  json
// @Param course_id path int true "Course ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /material/get [get]
// GetMaterials retrieves materials for a given course ID
func (server *Server) GetTotalMaterialInCourse(ctx *gin.Context) {
	var req GetTotalMaterialInCourseRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	materialCount, err := server.store.GetTotalMaterialsInCourse(ctx, req.CourseID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, materialCount)
}

// GetLessonCompletionRequest defines the request body structure for getting materials
type GetLessonCompletionRequest struct {
	UserID     int64 `form:"user_id"`
	CourseID   int64 `form:"course_id"`
	MaterialID int64 `form:"material_id"`
}

// @Summary Get materials for a course
// @Description Get materials for a course
// @ID get-materials
// @Accept  json
// @Produce  json
// @Param course_id path int true "Course ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /material/get [get]
// GetMaterials retrieves materials for a given course ID
func (server *Server) GetLessonCompletion(ctx *gin.Context) {
	var req GetLessonCompletionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	lesson, err := server.store.GetLessonCompletion(ctx, db.GetLessonCompletionParams{
		CourseID:   req.CourseID,
		UserID:     req.UserID,
		MaterialID: req.MaterialID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

type CreateLessonCompletionRequest struct {
	CourseID   int64 `json:"course_id"`
	UserID     int64 `json:"user_id"`
	MaterialID int64 `json:"material_id"`
}

// @Summary Create a new lesson completion
// @Description Create a new lesson completion
// @ID create-lesson_completion
// @Accept  json
// @Produce  json
// @Param request body CreateMaterialRequest true "Create Material Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /lessonCompletion [post]
// CreateLessonCompletion creates a new material
func (server Server) CreateLessonCompletion(ctx *gin.Context) {
	var req CreateLessonCompletionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.UserID != req.UserID {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	arg := db.CreateLessonCompletionTxParams{
		CreateLessonCompletionParams: db.CreateLessonCompletionParams{
			UserID:     req.UserID,
			CourseID:   req.CourseID,
			MaterialID: req.MaterialID,
			Completed:  true,
		},
		AfterCreate: func(LessonCompletion db.LessonCompletion) error {
			// Use Redis for task distribution
			taskPayload := &worker.PayloadCreateLessonCompletion{
				UserID:     LessonCompletion.UserID,
				CourseID:   LessonCompletion.CourseID,
				MaterialID: LessonCompletion.MaterialID,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskCreateLessonCompletion(ctx, taskPayload, opts...)
		},
	}

	txResult, err := server.store.CreateLessonCompletionTx(ctx, arg)

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

// course progreass controller
package api

import (
	"database/sql"
	db "eduApp/db/sqlc"
	"eduApp/token"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createCourseProgressRequest struct {
	CourseID int64 `json:"course_id"`
	UserID   int64 `json:"user_id"`
	Progress int64 `json:"progress"`
}

// @Summary Create course progress
// @Description Creates course progress
// @Tags CourseProgress
// @Accept json
// @Produce json
// @Param progress body createCourseProgressRequest true "Course progress object"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /createprogress [post]
// createCourseProgressRequest defines the request body structure for creating course progress
func (server *Server) createCourseProgress(ctx *gin.Context) {
	var req createCourseProgressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	courseProgress, err := server.store.CreateCourseProgress(ctx, db.CreateCourseProgressParams{
		CourseID: req.CourseID,
		UserID:   req.UserID,
		Progress: req.Progress,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, courseProgress)
}

type getCourseProgressRequest struct {
	CourseID int64 `form:"course_id"`
	UserID   int64 `form:"user_id"`
}

// @Summary Get course progress by ID
// @Description Returns course progress by ID
// @Tags CourseProgress
// @Accept json
// @Produce json
// @Param courseprogress_id path int true "Course progress ID"
// @Param enrolment_id query int true "Enrolment ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /courseProgress/get [get]
// getCourseProgress returns course progress by ID
func (server *Server) getCourseProgress(ctx *gin.Context) {
	var req getCourseProgressRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	courseProgress, err := server.store.GetCourseProgress(ctx, db.GetCourseProgressParams{
		CourseID: req.CourseID,
		UserID:   req.UserID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, courseProgress)
}

type listCourseProgressRequest struct {
	UserID   int64 `form:"user_id"`
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=100"`
}

// @Summary List course progress
// @Description Returns a list of course progress
// @Tags CourseProgress
// @Accept json
// @Produce json
// @Param enrolment_id query int true "Enrolment ID"
// @Param limit query int false "Limit the number of results"
// @Param offset query int false "Offset for paginated results"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /courseprogress/get [get]
// listCourseProgress returns a list of course progress
func (server *Server) ListCourseProgressByUser(ctx *gin.Context) {
	var req listCourseProgressRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListCourseProgressByUserParams{
		UserID: req.UserID,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	courseProgressList, err := server.store.ListCourseProgressByUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, courseProgressList)
}

// DeleteCourseprogressRequest defines the request body structure for deleting a course progress
type DeleteCourseprogressRequest struct {
	CourseprogressID int64 `form:"courseprogress_id"`
}

// @Summary Delete a Course Progress
// @Description Delete a course progress
// @ID delete-course progress
// @Accept  json
// @Produce  json
// @Param courseprogress_id path int true "CourseprogressID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /courseProgress/delete [delete]
// DeleteCourseProgress deletes a course progress
func (server *Server) DeleteCourseProgress(ctx *gin.Context) {
	var req DeleteCourseprogressRequest
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

	err := server.store.DeleteCourseProgress(ctx, req.CourseprogressID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Course Progress deleted successfully"})
}

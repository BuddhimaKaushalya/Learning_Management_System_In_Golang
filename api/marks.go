package api

import (
	db "eduApp/db/sqlc"
	"eduApp/token"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateMarkRequest defines the request body structure for creating a mark
type CreateMarkRequest struct {
	CourseID int64 `json:"course_id"`
	UserID   int64 `json:"user_id"`
	Marks    int64 `json:"marks"`
}

// @Summary Create a new mark
// @Description Create a new mark
// @ID create-mark
// @Accept json
// @Produce json
// @Param request body CreateMarkRequest true "Create Mark Request"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /create/mark [post]
// CreateMark creates a new mark
func (server *Server) CreateMark(ctx *gin.Context) {
	var req CreateMarkRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("authenticated users only able to make changes, access denied")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.CreateMarkParams{
		CourseID: req.CourseID,
		UserID:   req.UserID,
		Marks:    req.Marks,
	}

	mark, err := server.store.CreateMark(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, mark)
}

type DeleteMarkRequest struct {
	MarkID int64 `form:"mark_id"`
}

// @Summary Delete a mark
// @Description Delete a Mark by its MarkID
// @ID delete-mark
// @Accept json
// @Produce json
// @Param mark_id path int true "Mark ID"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /mark/delete [delete]
func (server *Server) DeleteMark(ctx *gin.Context) {
	var req DeleteMarkRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("not an admin of the system")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	err := server.store.DeleteMark(ctx, req.MarkID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Mark deleted successfully"})
}

type GetMarkRequest struct {
	MarkID int64 `form:"mark_id"`
}

// @Summary Get a mark by ID
// @Description Get a mark by its ID
// @ID get-mark
// @Accept json
// @Produce json
// @Param mark_id path int true "Mark ID"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /mark/get [get]
func (server *Server) GetMark(ctx *gin.Context) {
	var req GetMarkRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	mark, err := server.store.GetMark(ctx, req.MarkID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if authPayload.UserID != mark.UserID {
		err := errors.New("not authorized to access this mark")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, mark)
}

type ListMarksRequest struct {
	CourseID int64 `form:"course_id"`
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=100"`
}

// @Summary List marks for a course
// @Description List marks based on course ID
// @Produce json
// @Param course_id query int true "Course ID"
// @Param page_id query int true "Page ID"
// @Param page_size query int true "Page Size"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /marks/bycourse [get]
func (server *Server) ListMarks(ctx *gin.Context) {
	var req ListMarksRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("not an admin of the system")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	arg := db.ListMarksParams{
		CourseID: req.CourseID,
		Limit:    req.PageSize,
		Offset:   (req.PageID - 1) * req.PageSize,
	}

	marks, err := server.store.ListMarks(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, marks)
}

// UpdateMarkRequest defines the request body structure for updating a mark
type UpdateMarkRequest struct {
	MarkID   int64 `json:"mark_id" binding:"required"`
	Marks    int64 `json:"marks"`
	UserID   int64 `json:"user_id"`
	CourseID int64 `json:"course_id"`
}

// @Summary Update a mark
// @Description Update a mark
// @ID update-mark
// @Accept json
// @Produce json
// @Param mark_id path int true "Mark ID"
// @Param course_id path int true "Course ID"
// @Param request body UpdateMarkRequest true "Update Mark Request"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /mark/edit [put]
// UpdateMark updates a mark
func (server *Server) UpdateMark(ctx *gin.Context) {
	var req UpdateMarkRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("not an admin of the system")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	arg := db.UpdateMarkParams{
		UserID:   req.UserID,
		MarkID:   req.MarkID,
		Marks:    req.Marks,
		CourseID: req.CourseID,
	}

	mark, err := server.store.UpdateMark(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, mark)
}

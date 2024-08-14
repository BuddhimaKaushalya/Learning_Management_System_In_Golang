package api

import (
	db "eduApp/db/sqlc"
	"eduApp/token"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CareteCategoryrequest struct {
	Category string `json:"category"`
}

// @Summary Create a new category
// @Description Creates a new category
// @Accept multipart/form-data
// @Produce json
// @Param category_id formData int true "Category ID"
// @Param catagory formData string true "Category"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /category [post]
func (server *Server) CreateCategory(ctx *gin.Context) {
	var req CareteCategoryrequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	category, err := server.store.CreateCategory(ctx, req.Category)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, category)
}

type GetCategoryRequest struct {
	CategoryID int64 `form:"category_id"`
}

// @Summary Get an category by ID
// @ID get-category
// @Accept json
// @Produce json
// @Param category_id path int true "Category ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /category/get [get]
func (server *Server) GetCategory(ctx *gin.Context) {
	var req GetCategoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	category, err := server.store.GetCategory(ctx, req.CategoryID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, category)
}

// UpdateCategoryRequest contains the input parameters of update a Category
type UpdateCategoryRequest struct {
	CategoryID int64  `json:"category_id"`
	Category   string `json:"category"`
}

// @Summary Update Category
// @Description Update Category of certain category
// @Accept json
// @Produce json
// @Param request body UpdateCategoryRequest true "Update Category of certain category"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /category/edit [Patch]
// UpdateCategory function is api to call , to update a assignement from the db
func (server *Server) UpdateCategory(ctx *gin.Context) {
	var req UpdateCategoryRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.UpdateCategoryParams{
		CategoryID: req.CategoryID,
		Category:   req.Category,
	}
	category, err := server.store.UpdateCategory(ctx, db.UpdateCategoryParams(arg))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Update Category,  Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, category)
}

type DaleteCategoryRequest struct {
	CategoryID int64 `form:"category_id"`
}

// @Summary Delete a category
// @ID delete-category
// @Accept json
// @Produce json
// @Param category_id path int true "Category ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /category/delete [delete]
func (server *Server) DeleteCategory(ctx *gin.Context) {
	var req DaleteCategoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("you are not an admin of this system")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	err := server.store.DeleteCategory(ctx, req.CategoryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

type ListAllCategoriesRequest struct {
	PageID   int32 `form:"page_id,min=1"`
	PageSize int32 `form:"page_size,min=10,max=10"`
}

// @Summary List categories
// @Description Lists categories with pagination
// @Produce json
// @Param limit query int true "Number of items to return"
// @Param offset query int true "Offset for pagination"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /categories [get]
// ListAllCategoriesRequest defines the request body structure for listing categories
// ListAllCategories lists categories
func (server *Server) ListAllCategories(ctx *gin.Context) {
	var req ListAllCategoriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListAllCategoriesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	categories, err := server.store.ListAllCategories(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, categories)
}

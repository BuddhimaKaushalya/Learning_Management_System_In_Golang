package api

import (
	db "eduApp/db/sqlc"
	"eduApp/token"
	"eduApp/util"
	"eduApp/worker"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
)

// uploading a single file
func (server *Server) uploadSingleMaterial(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Convert file extension to lowercase
	fileExt := strings.ToLower(filepath.Ext(header.Filename))

	// Validate file extension (optional)
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".txt":  true,
		".docx": true,
		".pdf":  true,
		".mp4":  true,
		".jfif": true,
	}
	if !allowedExtensions[fileExt] {
		return "", fmt.Errorf("unsupported file extension: %s", fileExt)
	}

	// Generate unique filename
	originalFileName := strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename))
	now := time.Now()
	filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt

	// Create upload directory if it doesn't exist
	uploadDir := "uploads/materials"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Create destination file path
	filePath := filepath.Join(uploadDir, filename)
	// Save uploaded file
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return "", err
	}

	// URL is split into parts
	parts := strings.Split(filePath, "/")
	println(parts)

	// Replace "uploads" with "static"
	if len(parts) > 2 && parts[0] == "uploads" {
		parts[0] = "static"
	}
	println(parts)

	//newUrl for load file
	newUrl := strings.Join(parts, "/")
	println("LINK", newUrl)

	serverAddress := server.config.FileSource
	print("server address", serverAddress)
	MaterialFile := fmt.Sprintf("%s/%s", serverAddress, newUrl)
	println(MaterialFile)

	return MaterialFile, nil
}

// CreateMaterialRequest defines the request body structure for creating a material
type CreateMaterialRequest struct {
	CourseID     int64  `form:"course_id"`
	Title        string `form:"title"`
	MaterialFile string `json:"material_file"`
	OrderNumber  int64  `form:"order_number"`
}

// @Summary Create a new material
// @Description Create a new material
// @ID create-material
// @Accept  json
// @Produce  json
// @Param request body CreateMaterialRequest true "Create Material Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /material [post]
// CreateMaterial creates a new material
func (server *Server) CreateMaterial(ctx *gin.Context) {
	var req CreateMaterialRequest

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

	var materialFile string
	file, header, err := ctx.Request.FormFile("material_file")
	if err == nil {
		materialFile, err = server.uploadSingleMaterial(file, header)
		println("material_file", materialFile)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	req.MaterialFile = materialFile

	arg := db.CreateMaterialTxParams{
		CreateMaterialParams: db.CreateMaterialParams{
			CourseID:     req.CourseID,
			Title:        req.Title,
			MaterialFile: materialFile,
			OrderNumber:  req.OrderNumber,
		},
		AfterCreate: func(material db.Material) error {
			// Use Redis for task distribution
			taskPayload := &worker.PayloadCreateMaterials{
				MaterialID:   material.MaterialID,
				GetPayload:   authPayload,
				Title:        req.Title,
				MaterialFile: materialFile,
				OrderNumber:  req.OrderNumber,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskCreateMaterials(ctx, taskPayload, opts...)
		},
	}

	txResult, err := server.store.CreateMaterialTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create material, Please try again."})
		return
	}

	ctx.JSON(http.StatusOK, txResult)
}

// GetMaterialRequest contains the input parameters for Get material
type GetMaterialRequest struct {
	CourseID   int64 `form:"course_id"`
	MaterialID int64 `form:"material_id"`
}

// @Summary Get Material
// @Description Get Material of certain course
// @Accept json
// @Produce json
// @Param request body DeleteMaterialRequest true "Get Material of certain course"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /material/get [Get]
// GetMaterial function is api call to get material from db related to certain course
func (server *Server) GetMaterial(ctx *gin.Context) {
	var req GetMaterialRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	subscription, err := server.store.GetSubscriptionByUser(ctx, db.GetSubscriptionByUserParams{
		UserID:   authPayload.UserID,
		CourseID: req.CourseID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if !subscription.Active {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user doesn't have an active subscription"})
		return
	}

	arg := db.GetMaterialParams{
		CourseID:   req.CourseID,
		MaterialID: req.MaterialID,
	}

	material, err := server.store.GetMaterial(ctx, db.GetMaterialParams(arg))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Check if the previous material is completed
	previousMaterialOrderNumber := material.OrderNumber - 1
	if previousMaterialOrderNumber > 0 {
		previousMaterial, err := server.store.GetMaterialByOrderNumber(ctx, db.GetMaterialByOrderNumberParams{
			CourseID:    req.CourseID,
			OrderNumber: previousMaterialOrderNumber,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		previousLessonCompleted, _ := server.store.GetLessonCompletion(ctx, db.GetLessonCompletionParams{
			UserID:     authPayload.UserID,
			CourseID:   req.CourseID,
			MaterialID: previousMaterial.MaterialID,
		})

		if !previousLessonCompleted.Completed {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "you haven't completed the previous material yet"})
			return
		}
	}

	ctx.JSON(http.StatusOK, material)
}

// UpdateMaterialRequest defines the request body structure for updating a material
type UpdateMaterialRequest struct {
	MaterialID   int64  `form:"material_id"`
	Title        string `form:"title"`
	MaterialFile string `json:"material_file"`
	CourseID     int64  `form:"course_id"`
}

// @Summary Update a material
// @Description Update a material
// @ID update-material
// @Accept  json
// @Produce  json
// @Param material_id path int true "Material ID"
// @Param request body UpdateMaterialRequest true "Update Material Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /material/edit [put]
// UpdateMaterial updates a material
func (server *Server) UpdateMaterial(ctx *gin.Context) {
	var req UpdateMaterialRequest

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

	// Handle image upload (if included in the request)
	var materialF string
	file, header, err := ctx.Request.FormFile("material_file")
	if err == nil {
		materialF, err = server.uploadSingleFile(file, header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	// Update request struct with image filename (if uploaded)
	req.MaterialFile = materialF

	//previous details to remove file
	getMaterial, err := server.store.GetMaterial(ctx, db.GetMaterialParams{
		MaterialID: req.MaterialID,
		CourseID:   req.CourseID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//assign value to the file path
	value := getMaterial.MaterialFile
	//remove exiting fil
	util.DeleteFileByURL(value)

	//pass new args
	arg := db.UpdateMaterialParams{
		MaterialID:   req.MaterialID,
		Title:        pgtype.Text{String: req.Title, Valid: true},
		MaterialFile: pgtype.Text{String: materialF, Valid: true},
	}

	material, err := server.store.UpdateMaterial(ctx, db.UpdateMaterialParams(arg))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Update Material,  Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, material)

}

// DeleteMaterialRequest defines the request body structure for deleting a material
type DeleteMaterialRequest struct {
	MaterialID int64 `form:"material_id"`
}

// @Summary Delete a material
// @Description Delete a material
// @ID delete-material
// @Accept  json
// @Produce  json
// @Param material_id path int true "Material ID"
// @Param course_id path int true "Course ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /material/delete [delete]
// DeleteMaterial deletes a material
func (server *Server) DeleteMaterial(ctx *gin.Context) {
	var req DeleteMaterialRequest
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

	err := server.store.DeleteMaterial(ctx, req.MaterialID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Material deleted successfully"})
}

// ListMaterialsByCourse contains the input parameters for list course details
type ListMaterialsByCourse struct {
	CourseID int64 `form:"course_id"`
}

// @Summary List Material
// @Description List Material of certain course
// @Accept json
// @Produce json
// @Param request body ListMaterialsByCourse true "List Material of certain course"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /material/list [Get]
// ListMaterial is a funtion to list material details from the database
func (server *Server) ListMaterial(ctx *gin.Context) {
	var req ListMaterialsByCourse

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	material, err := server.store.ListMaterial(ctx, req.CourseID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list material data , Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, material)
}

// GetTotalMaterialInCourseRequest defines the request body structure for getting materials
type GetTotalMaterialInCourseRequest struct {
	CourseID int64 `form:"course_id,min=1"`
}

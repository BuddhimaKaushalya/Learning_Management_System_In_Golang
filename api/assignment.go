package api

import (
	db "eduApp/db/sqlc"
	"eduApp/token"
	"eduApp/util"
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
)

// uploading a single file
func (server *Server) uploadSingleFile(file multipart.File, header *multipart.FileHeader) (string, error) {

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
	uploadDir := "uploads/assignments"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Create destination file path
	filePath := filepath.Join(uploadDir, filename)
	println(filePath)

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

	// Split the URL into parts
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
	files := fmt.Sprintf("%s/%s", serverAddress, newUrl)
	println(files)

	return files, nil
}

type createAssignmentRequest struct {
	Title          string `form:"title"`
	CourseID       int64  `form:"course_id"`
	DueDate        string `form:"due_date"`
	AssignmentFile string `json:"assignment_file"`
}

// @Summary Create a new assignment
// @Description Create a new assignment with the provided details
// @ID create-assignment
// @Accept json
// @Produce json
// @Param request body createAssignmentRequest true "Assignment details"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /assignments [post]
func (server *Server) CreateAssignment(ctx *gin.Context) {
	var req createAssignmentRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Handle file upload (if included in the request)
	var assignmentFile string
	file, header, err := ctx.Request.FormFile("assignment_file")
	if err == nil {
		assignmentFile, err = server.uploadSingleFile(file, header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed: " + err.Error()})
			return
		}
	}

	// Update request struct with filename (if uploaded)
	req.AssignmentFile = assignmentFile

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due date format: " + err.Error()})
		return
	}

	arg := db.CreateAssignmentParams{
		Title:          req.Title,
		DueDate:        req.DueDate,
		AssignmentFile: assignmentFile,
		CourseID:       req.CourseID,
	}

	assignment, err := server.store.CreateAssignment(ctx, db.CreateAssignmentParams(arg))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, assignment)
}

type GetAssignmentRequest struct {
	AssignmentID int64 `form:"assignment_id"`
}

// @Summary Get an assignment by ID
// @ID get-assignment
// @Accept json
// @Produce json
// @Param assignment_id path int true "Assignment ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /assignment/get [get]
func (server *Server) getAssignment(ctx *gin.Context) {
	var req GetAssignmentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	assignment, err := server.store.GetAssignment(ctx, req.AssignmentID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, assignment)
}

// UpdateAssignmentReques contains the input parameters of update a Assignment
type UpdateAssignmentRequest struct {
	AssignmentID   int64  `form:"assignment_id"`
	Title          string `form:"title"`
	DueDate        string `form:"due_date"`
	AssignmentFile string `json:"assignment_file"`
	CourseID       int64  `form:"course_id"`
}

// @Summary Update Asssignment
// @Description Update Asssignment of certain material
// @Accept json
// @Produce json
// @Param request body UpdateAssignmentRequest true "Update Asssignment of certain material"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /assignments/edit [Patch]
// UpdateAssignment function is api to call , to update a assignement from the db
func (server *Server) UpdateAssignment(ctx *gin.Context) {
	var req UpdateAssignmentRequest

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
	var assignmentF string
	file, header, err := ctx.Request.FormFile("assignment_file")
	if err == nil {
		assignmentF, err = server.uploadSingleFile(file, header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}
	// Update request struct with image filename (if uploaded)
	req.AssignmentFile = assignmentF

	//previous details to remove file
	getAssignment, err := server.store.GetAssignment(ctx, req.AssignmentID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	//assign file path
	value := getAssignment.AssignmentFile
	//remove exiting file
	util.DeleteFileByURL(value)

	arg := db.UpdateAssignmentParams{
		AssignmentID:   req.AssignmentID,
		Title:          req.Title,
		DueDate:        req.DueDate,
		AssignmentFile: req.AssignmentFile,
		CourseID:       req.CourseID,
	}
	assignment, err := server.store.UpdateAssignment(ctx, db.UpdateAssignmentParams(arg))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Update Assignment,  Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, assignment)
}

type DeleteAssignmentRequest struct {
	AssignmentID int64 `form:"assignment_id"`
}

// @Summary Delete an assignment
// @ID delete-assignment
// @Accept json
// @Produce json
// @Param assignment_id path int true "Assignment ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /assignment/delete [delete]
func (server *Server) deleteAssignment(ctx *gin.Context) {
	var req DeleteAssignmentRequest
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

	err := server.store.DeleteAssignment(ctx, req.AssignmentID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Assignment deleted successfully"})
}

package api

import (
	"database/sql"
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
func (server *Server) uploadResource(file multipart.File, header *multipart.FileHeader) (string, error) {

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
		"jfif":  true,
	}
	if !allowedExtensions[fileExt] {
		return "", fmt.Errorf("unsupported file extension: %s", fileExt)
	}

	// Generate unique filename
	originalFileName := strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename))
	now := time.Now()
	filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt

	// Create upload directory if it doesn't exist
	uploadDir := "uploads/submissions"
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

	serverAddress := server.config.FileSource
	File := fmt.Sprintf("%s/%s", serverAddress, filePath)

	return File, nil
}

// CreateSubmissionRequest defines the request body structure for creating a submission
type CreateSubmissionRequest struct {
	AssignmentID     int64  `form:"assignment_id"`
	UserID           int64  `form:"user_id"`
	Grade            string `form:"grade"`
	Resource         string `json:"resource"`
	DateOfSubmission string `form:"date_of_submission"`
	Submitted        bool   `form:"submitted"`
}

// @Summary Create a new Submission
// @Description Creates a new submission by depend on assignment
// @Accept json
// @Produce json
// @Param request body CreateSubmissionRequest true "assignmnet_id and user_id"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /submission/create [post]
// CreateSubmission creates a new submission
func (server *Server) CreateSubmission(ctx *gin.Context) {
	var req CreateSubmissionRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Handle image upload (if included in the request)
	var resourceFile string
	file, header, err := ctx.Request.FormFile("resource")
	if err == nil {
		resourceFile, err = server.uploadResource(file, header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	// Update request struct with filename (if uploaded)
	req.Resource = resourceFile

	arg := db.CreateSubmissionParams{
		AssignmentID:     req.AssignmentID,
		UserID:           req.UserID,
		Submitted:        true,
		Grade:            "null",
		DateOfSubmission: time.Now(),
		Resource:         resourceFile,
	}

	submission, err := server.store.CreateSubmission(ctx, db.CreateSubmissionParams(arg))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

// GetsubmissionsByAssignmentRequest defines the request body structure for getting a submission
type GetsubmissionsByAssignmentRequest struct {
	AssignmentID int64 `form:"assignment_id"`
}

// @Summary Get a submission By Assignment
// @Description Retrieves a submission by assignment and student ID
// @ID GetSubmissionsByAssignment
// @Produce json
// @Param assignment_id path int true "Assignment ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /submissions/byassignemnt [get]
// GetSubmissionsByAssignment retrieves a submission by assignment
func (server *Server) GetSubmissionsByAssignment(ctx *gin.Context) {
	var req GetsubmissionsByAssignmentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	submission, err := server.store.GetsubmissionsByAssignment(ctx, req.AssignmentID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

// GetsubmissionsByUserRequest defines the request body structure for getting a submission
type GetsubmissionsByUserRequest struct {
	UserID int64 `form:"user_id"`
}

// @Summary Get a submission By User
// @Description Retrieves a submission by assignment and student ID
// @ID GetSubmissionsByUser
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /submissions/byuser [get]
// GetSubmissionsByUser retrieves a submission by user
func (server *Server) GetSubmissionsByUser(ctx *gin.Context) {
	var req GetsubmissionsByUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" || authPayload.UserID == req.UserID {
		err := errors.New("you are not an authorized user")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}
	submission, err := server.store.GetsubmissionsByUser(ctx, req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

// listSubmissionsRequest defines the request body structure for listing submissions
type listSubmissionsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=100"`
}

// @Summary List submissions
// @Description Lists submissions for a given assignment ID
// @ID listSubmissions
// @Produce json
// @Param assignment_id path int true "Assignment ID"
// @Param limit query int true "Limit" minimum(1) maximum(100)
// @Param offset query int true "Offset" minimum(0)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /submissions [get]
// listSubmissions lists submissions
func (server *Server) listSubmissions(ctx *gin.Context) {
	var req listSubmissionsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("you are not an authorized user")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	arg := db.ListsubmissionsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	submissions, err := server.store.Listsubmissions(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, submissions)
}

// UpdateSubmissionRequest defines the request body structure for updating a submission
type UpdateSubmissionRequest struct {
	SubmissionID int64 `json:"submission_id"`
	AssignmentID int64 `json:"assignment_id"`
	UserID       int64 `json:"user_id"`
}

// @Summary Update a submission
// @ID update-submission
// @Accept  json
// @Produce  json
// @Param material_id path int true "Material ID"
// @Param request body UpdateSubmissionRequest true "Update Submission Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /submission/edit [put]
// UpdateSubmission updates a submission
func (server *Server) UpdateSubmission(ctx *gin.Context) {
	var req UpdateSubmissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if req.UserID != authPayload.UserID {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	arg := db.UpdateSubmissionParams{
		UserID:       req.UserID,
		AssignmentID: req.AssignmentID,
		SubmissionID: req.SubmissionID,
	}

	submission, err := server.store.UpdateSubmission(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, submission)
}

// DeleteSubmissionRequest defines the request body structure for deleting a submission
type DeleteSubmissionRequest struct {
	AssignmentID int64 `form:"assignment_id"`
	UserID       int64 `form:"user_id"`
}

// @Summary Delete a submission
// @Description Delete a submission
// @ID delete-submission
// @Accept  json
// @Produce  json
// @Param submission_id path int true "Submission ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /submission/delete [delete]
// DeleteSubmission deletes a submission
func (server *Server) DeleteSubmission(ctx *gin.Context) {
	var req DeleteSubmissionRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.UserID == req.UserID {
		err := errors.New("you are not an authorized user")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	arg := db.GetSubmissionParams{
		AssignmentID: req.AssignmentID,
		UserID:       req.UserID,
	}

	getSubmission, err := server.store.GetSubmission(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	//assign value to file
	value := getSubmission.Resource
	//remove the exiting file
	util.DeleteFileByURL(value)

	//delete submission details
	errors := server.store.DeleteSubmission(ctx, db.DeleteSubmissionParams{
		AssignmentID: req.AssignmentID,
		UserID:       req.UserID,
	})
	if errors != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Submission data Deletion Failed, Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"messege": "Submission successfully deleted.!"})
}

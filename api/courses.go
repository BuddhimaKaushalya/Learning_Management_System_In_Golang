package api

import (
	"database/sql"
	db "eduApp/db/sqlc"
	"eduApp/token"
	"eduApp/typetext"
	"eduApp/util"
	"eduApp/worker"
	"encoding/json"
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
func (server *Server) uploadSingleImage(file multipart.File, header *multipart.FileHeader) (string, error) {

	// Convert file extension to lowercase
	fileExt := strings.ToLower(filepath.Ext(header.Filename))

	// Validate file extension (optional)
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
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
	uploadDir := "uploads/course"
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

	// Split the URL into parts
	parts := strings.Split(filePath, "/")
	println(parts)

	// Replace "uploads" with "static"
	if len(parts) > 2 && parts[0] == "uploads" {
		parts[0] = "static"
	}
	println(parts)

	//newUrl for load images
	newUrl := strings.Join(parts, "/")
	println("LINK", newUrl)

	serverAddress := server.config.FileSource
	print("server address", serverAddress)
	Image := fmt.Sprintf("%s/%s", serverAddress, newUrl)
	println(Image)
	return Image, nil

}

// CreateCourseRequest defines the request body structure for creating a course
type CreateCourseRequest struct {
	UserID           int64             `form:"user_id"`
	Title            string            `form:"title"`
	Description      string            `form:"description"`
	Image            string            `json:"image"`
	Catagory         string            `form:"catagory"`
	WhatWill         typetext.WhatWill `form:"what_will"`
	SequentialAccess bool              `form:"sequential_access"`
}

// @Summary Create a new course
// @Description Creates a new course
// @Accept multipart/form-data
// @Produce json
// @Param user_id formData int true "User ID"
// @Param title formData string true "Title"
// @Param description formData string true "Description"
// @Param catagory formData string true "Category"
// @Param sequential_access formData bool true "Sequential Access"
// @Param what_will formData string true "What Will"
// @Param image formData file false "Profile Picture"
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @Router /course [post]
func (server *Server) CreateCourse(ctx *gin.Context) {
	var req CreateCourseRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	println("userID1", req.UserID)

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" || authPayload.UserID != req.UserID {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	println("userID2", authPayload.Role)

	// Handle image upload (if included in the request)
	var imageFile string
	file, header, err := ctx.Request.FormFile("image")
	if err == nil {
		imageFile, err = server.uploadSingleImage(file, header)
		println("image_files", imageFile)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	// Update request struct with image filename (if uploaded)
	req.Image = imageFile

	arg := db.CreateCoursesParams{
		UserID:           req.UserID,
		Title:            req.Title,
		Description:      req.Description,
		Image:            imageFile,
		Catagory:         req.Catagory,
		SequentialAccess: true,
		WhatWill:         req.WhatWill,
	}

	course, err := server.store.CreateCourses(ctx, db.CreateCoursesParams(arg))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create course: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, course)
}

// GetCourseRequest defines the request body structure for getting a course
type GetCourseRequest struct {
	CourseID int64 `form:"course_id,min=1"`
}

// // @Summary Get a course by ID
// // @Description Retrieves a course by its ID
// // @Produce json
// // @Param course_id path int true "Course ID"
// // @Success 200
// // @Failure 400
// // @Failure 404
// // @Failure 500
// // @Router /course/get [get]
// // GetCourse retrieves a course by ID
func (server *Server) GetCourse(ctx *gin.Context) {
	var req GetCourseRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	course, err := server.store.GetCourses(ctx, req.CourseID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, course)
}

type ListCoursesRequest struct {
	PageID   int32 `form:"page_id,min=1"`
	PageSize int32 `form:"page_size,min=10,max=10"`
}

// @Summary List courses
// @Description Lists courses with pagination
// @Produce json
// @Param limit query int true "Number of items to return"
// @Param offset query int true "Offset for pagination"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /courses [get]
// ListCoursesRequest defines the request body structure for listing courses
// ListCourses lists courses
func (server *Server) ListCourses(ctx *gin.Context) {
	var req ListCoursesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListCoursesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	courses, err := server.store.ListCourses(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, courses)
}

// NullString is a wrapper around sql.NullString that provides JSON serialization/deserialization.
type NullString struct {
	sql.NullString
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		ns.String = *s
		ns.Valid = true
	} else {
		ns.String = ""
		ns.Valid = false
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(ns.String)
}

// UpdateCoursesRequest contains the input parameters for update course data
type UpdateCoursesRequest struct {
	Title            string            `form:"title"`
	Image            string            `json:"image"`
	Description      string            `form:"description"`
	Catagory         string            `form:"catagory"`
	SequentialAccess bool              `form:"sequential_access"`
	CourseID         int64             `form:"course_id"`
	WhatWill         typetext.WhatWill `form:"what_will"`
}

// @Summary Update Course details
// @Description Update Course details
// @Accept json
// @Produce json
// @Param request body UpdateCoursesRequest true "Update Course details "
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /course/edit [Patch]
// UpdateCourse funtion is api call to update course data in db
func (server *Server) UpdateCourse(ctx *gin.Context) {
	var req UpdateCoursesRequest

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
	var updatedImage string
	file, header, err := ctx.Request.FormFile("image")
	if err == nil {
		updatedImage, err = server.uploadSingleImage(file, header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	// Update request struct with image filename (if uploaded)
	req.Image = updatedImage

	//get course for remove exiting image
	getCourse, err := server.store.GetCourses(ctx, req.CourseID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	//assign value to the image path
	value := getCourse.Image

	//remove exiting file
	util.DeleteFileByURL(value)

	arg := db.UpdateCoursesParams{
		Title:            pgtype.Text{String: req.Title, Valid: true},
		CourseID:         req.CourseID,
		Description:      pgtype.Text{String: req.Description, Valid: true},
		Image:            pgtype.Text{String: req.Image, Valid: true},
		Catagory:         pgtype.Text{String: req.Catagory, Valid: true},
		SequentialAccess: pgtype.Bool{Bool: true, Valid: true},
		WhatWill:         req.WhatWill,
	}

	course, err := server.store.UpdateCourses(ctx, db.UpdateCoursesParams(arg))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Update course data, Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, course)
}

// deleteCourseRequest defines the request body structure for deleting a Course
type deleteCourseRequest struct {
	CourseID int64 `form:"course_id"`
}

// @Summary Delete a course
// @Description Deletes a course by ID
// @Produce json
// @Param course_id query int true "Course ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /course/delete [delete]
// DeleteCourse deletes a Course
// deleteCourseRequest defines the request body structure for deleting a course

func (server *Server) DeleteCourse(ctx *gin.Context) {
	var req deleteCourseRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Enqueue the delete task
	taskPayload := &worker.PayloadDeleteCourse{
		CourseID: req.CourseID,
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

	// Enqueue the delete task
	err := server.taskDistributor.DistributeTaskDeleteCourse(ctx, taskPayload, opts...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Return immediately after enqueuing the task
	ctx.JSON(http.StatusOK, gin.H{"message": "Course deletion task enqueued successfully"})
}

// Get student Count
func (server *Server) GetCourseCount(ctx *gin.Context) {
	var req StudenTCountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	count, err := server.store.StudentCount(ctx, req.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"student_count": count})
}

// ListAllCourseByCatagoryRequest contains input parameters for list course data by created user
type ListAllCourseByCatagoryRequest struct {
	Catagory string `form:"catagory"`
}

// @Summary List Course details  by catagory
// @Description List Course details by catagory
// @Accept json
// @Produce json
// @Param request body ListAllCourseByCatagoryRequest true "List Course details by catagory"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /list/course/bycatagory [Get]
// ListCourseByCreatedUser funtion is api call to list course from db by catagory
func (server *Server) ListAllCourseByCatagory(ctx *gin.Context) {
	var req ListAllCourseByCatagoryRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := ListAllCourseByCatagoryRequest{
		Catagory: req.Catagory,
	}

	course, err := server.store.ListAllCourseByCatagory(ctx, arg.Catagory)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list all course data, Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, course)
}

// @Summary List Course details  by created user
// @Description List Course details by created user
// @Accept json
// @Produce json
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /list/catagories [Get]
// ListAllCourseCatagories funtion is api call to list course from db by created user
func (server *Server) ListAllCourseCatagories(ctx *gin.Context) {
	course, err := server.store.ListAllCourseCatagories(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list all course data, Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, course)
}

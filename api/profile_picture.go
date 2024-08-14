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
	"github.com/jackc/pgx/v5/pgtype"
)

// uploadSingleFile handles uploading a single file
func (server *Server) uploadSingleProfileImage(file multipart.File, header *multipart.FileHeader) (string, error) {

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
	uploadDir := "uploads/profile"
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

type CreateProfilePictureRequest struct {
	UserID  int64  `form:"user_id"`
	Picture string `json:"picture"`
}

// @Summary Create a new profile picture
// @Description Create a new profile picture with the provided details
// @ID create-profile picture
// @Accept json
// @Produce json
// @Param request body CreateProfilePictureRequest true "Profile picture details"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /profile [post]

func (server *Server) CreateProfilePicture(ctx *gin.Context) {
	var req CreateProfilePictureRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.UserID != int64(req.UserID) {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Handle image upload (if included in the request)
	var profileImageFile string
	file, header, err := ctx.Request.FormFile("picture")
	if err == nil {
		profileImageFile, err = server.uploadSingleProfileImage(file, header)
		println("image_files", profileImageFile)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed: " + err.Error()})
			return
		}
	}

	req.Picture = profileImageFile

	arg := db.CreateProfilePictureParams{
		UserID:  int64(req.UserID),
		Picture: profileImageFile,
	}

	profile, err := server.store.CreateProfilePicture(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, profile)
}

type UpdateProfilePictureRequest struct {
	UserID  int64  `form:"user_id"`
	Picture string `json:"picture"`
}

// @Summary Update Profile Picture
// @Description Update Profile Picture of a user
// @Accept json
// @Produce json
// @Param request body UpdateProfilePictureRequest true "Update Profile picture of a user"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /profilepicture/edit [Patch]
// UpdateProfilePicture function is api to call , to update a profile picture from the db

func (server *Server) UpdateProfilePicture(ctx *gin.Context) {
	var req UpdateProfilePictureRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.UserID != int64(req.UserID) {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Handle image upload (if included in the request)
	var profileImageFile string
	file, header, err := ctx.Request.FormFile("image")
	if err == nil {
		profileImageFile, err = server.uploadSingleProfileImage(file, header)
		println("image_files", profileImageFile)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}

	getPPImage, err := server.store.GetProfilePicture(ctx, req.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	//assign url value to var
	value := getPPImage.Picture

	//remove exiting file
	util.DeleteFileByURL(value)

	arg := db.UpdateProfilePictureParams{
		UserID: req.UserID,
		Picture: pgtype.Text{
			String: profileImageFile,
			Valid:  profileImageFile != "",
		},
	}

	profileImage, err := server.store.UpdateProfilePicture(ctx, db.UpdateProfilePictureParams(arg))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, profileImage)
}

type DeleteProfilePictureRequest struct {
	UserID int64 `form:"user_id"`
}

// @Summary Delete a profile picture
// @ID delete-profile
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /profile/delete[delete]

func (server *Server) DeleteProfilePicture(ctx *gin.Context) {
	var req DeleteProfilePictureRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.UserID != req.UserID {
		err := errors.New("authenticated users only able to make changes, access denied ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	getPPImage, err := server.store.GetProfilePicture(ctx, req.UserID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//assign url value to var
	value := getPPImage.Picture

	//remove exiting file
	util.DeleteFileByURL(value)

	errors := server.store.DeleteProfilePicture(ctx, req.UserID)

	if errors != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"messege": "Profile picture has been deleted succefully"})
}

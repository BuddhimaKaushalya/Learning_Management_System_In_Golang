package api

import (
	db "eduApp/db/sqlc"
	"eduApp/token"
	"eduApp/util"
	"eduApp/val"
	"eduApp/worker"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
)

type createUserRequest struct {
	UserName       string `json:"user_name" binding:"required,alphanum"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	HashedPassword string `json:"hashed_password" binding:"required,min=6"`
	Email          string `json:"email" binding:"required,email"`
	Role           string `json:"role"`
}

type userResponse struct {
	Username  string    `json:"user_name"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:  user.UserName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

// @Summary Create a new user
// @Description Create a new user with the provided details
// @ID create-user
// @Accept  json
// @Produce  json
// @Param request body createUserRequest true "User creation request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /signup [post]

func (server Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			UserName:       req.UserName,
			FirstName:      req.FirstName,
			LastName:       req.LastName,
			HashedPassword: hashedPassword,
			Email:          req.Email,
			Role:           "student",
		},
		AfterCreate: func(user db.User) error {
			// Use Redis for task distribution
			taskPayload := &worker.PayloadSendVerifyEmail{
				UserName: user.UserName,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	txResult, err := server.store.CreateUserTx(ctx, arg)

	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolations {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Respond to the client
	rsp := newUserResponse(txResult.User)
	ctx.JSON(http.StatusCreated, rsp)
}

type UpdateUserRequest struct {
	HashedPassword    NullString `json:"hashed_password"`
	PasswordChangedAt time.Time  `json:"password_changed_at"`
	FirstName         NullString `json:"first_name"`
	LastName          NullString `json:"last_name"`
	Email             NullString `json:"email"`
	IsEmailVerified   bool       `json:"is_email_verified"`
	UserName          NullString `json:"user_name"`
	UserID            int64      `json:"user_id"`
	Role              NullString `json:"role"`
}

// @Summary Update a user
// @Description Updates a user with provided details
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest true "Updated user details"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /upateUser [PATCH]
// Updateuser updates the selected user
// UpdateUser updates the selected user by UserID
func (server Server) UpdateUser(ctx *gin.Context) {
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if req.UserID != authPayload.UserID {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.HashedPassword.String)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := &db.UpdateUserParams{
		HashedPassword: pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		},
		PasswordChangedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		FirstName: pgtype.Text{
			String: req.FirstName.String,
			Valid:  true,
		},
		LastName: pgtype.Text{
			String: req.LastName.String,
			Valid:  true,
		},
		Email: pgtype.Text{
			String: req.Email.String,
			Valid:  true,
		},
		IsEmailVerified: pgtype.Bool{
			Bool:  req.IsEmailVerified,
			Valid: true,
		},
		UserName: pgtype.Text{
			String: req.UserName.String,
			Valid:  true,
		},
		Role: pgtype.Text{
			String: req.Role.String,
			Valid:  true,
		},
		UserID: req.UserID,
	}

	user, err := server.store.UpdateUser(ctx, db.UpdateUserParams{
		HashedPassword:    arg.HashedPassword,
		PasswordChangedAt: arg.PasswordChangedAt,
		FirstName:         arg.FirstName,
		LastName:          arg.LastName,
		Email:             arg.Email,
		IsEmailVerified:   arg.IsEmailVerified,
		UserName:          arg.UserName,
		UserID:            req.UserID,
		Role:              arg.Role,
	})
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolations {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	rsp := newUserResponse(user)

	ctx.JSON(http.StatusOK, rsp)
}

// UpdateUsersPasswordRequest contains the input parameters of update a User's pasword
type UpdateUsersPasswordRequest struct {
	CurrentPassword   string    `json:"current_password" `
	HashedPassword    string    `json:"hashed_password" `
	ConfirmPassword   string    `json:"confirm_password"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Email             string    `json:"email"`
}

// @Summary Update a user password
// @Description Update a user password with provided details
// @Accept json
// @Produce json
// @Param request body UpdateUsersPasswordRequest true "Update user's pasword request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /reset/password [Patch]
// UpdateUserPassword function is api to call , to update a user's account password in the db
func (server *Server) UpdateUserPassword(ctx *gin.Context) {
	var req UpdateUsersPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := val.ValidatePassword(req.HashedPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashPassword, err := util.HashPassword(req.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	checkPassword := util.CheckPassword(req.ConfirmPassword, hashPassword)
	if checkPassword != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Password mismatch, Please Try again.!"})
		return
	}

	arg := db.UpdateUsersPasswordParams{
		Email: req.Email,
		HashedPassword: pgtype.Text{
			String: hashPassword,
			Valid:  true,
		},
		PasswordChangedAt: pgtype.Timestamptz{
			Time:  req.PasswordChangedAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  req.UpdatedAt,
			Valid: true,
		},
	}

	_, err = server.store.UpdateUsersPassword(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password, Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"messege": "Reset Password is Successfull.!"})
}

// CheckEmailRequest contains the input parameters of check a User's email
type CheckEmailRequest struct {
	Email string `json:"email"`
}

// @Summary Check email
// @Description Check email for password reset process
// @Accept json
// @Produce json
// @Param request body CheckEmailRequest true "Check email request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /check/email [Get]
// CheckEmail function is api to call , to check a user's email exists in the db of not
func (server *Server) CheckEmail(ctx *gin.Context) {
	var req CheckEmailRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CheckEmailTxParams{
		Email: req.Email,
		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadResetPassword{
				Email: user.Email,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(5 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributorTaskResetpassword(ctx, taskPayload, opts...)
		},
	}

	_, err := server.store.CheckEmailTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Invalid email address, Please try again .!"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"messege": "Email found"})
}

type loginUserRequest struct {
	UserName       string `json:"user_name"`
	HashedPassword string `json:"hashed_password"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

// @Summary Log in user
// @Description Log in a user with the provided credentials
// @ID login-user
// @Accept  json
// @Produce  json
// @Param request body loginUserRequest true "Login request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /login [post]
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.UserName)

	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.HashedPassword, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.UserName,
		user.Role,
		user.UserID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.UserName,
		user.Role,
		user.UserID,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		SessionID: pgtype.UUID{
			Bytes: refreshPayload.ID,
			Valid: true,
		},
		UserID:       user.UserID,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		SessionID:             uuid.UUID(session.SessionID.Bytes),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, rsp)
}

type createAdminRequest struct {
	UserName       string `json:"user_name" binding:"required,alphanum"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	HashedPassword string `json:"hashed_password" binding:"required,min=6"`
	Email          string `json:"email" binding:"required,email"`
	Role           string `json:"role"`
}

type userAdminResponse struct {
	Username  string    `json:"user_name"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func newAdminResponse(user db.User) userAdminResponse {
	return userAdminResponse{
		Username:  user.UserName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

// @Summary Create a admin's user
// @Description Create a new admin's user with the provided details
// @ID create admin user
// @Accept  json
// @Produce  json
// @Param request body createUserRequest true "User creation request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /admin/signup [post]
func (server *Server) createAdminUser(ctx *gin.Context) {
	var req createAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			UserName:       req.UserName,
			FirstName:      req.FirstName,
			LastName:       req.LastName,
			HashedPassword: hashedPassword,
			Email:          req.Email,
			Role:           "admin",
		},
		AfterCreate: func(user db.User) error {
			// Use Redis for task distribution
			taskPayload := &worker.PayloadSendVerifyEmail{
				UserName: user.UserName,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	user, err := server.store.CreateUserTx(ctx, arg)

	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolations {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newAdminResponse(user.User)
	ctx.JSON(http.StatusCreated, rsp)
}

// ListUserRequest contains the impurt parameters for list rolsbased user data
type ListUserRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// @Summary ListUser
// @Description ListUser with the provided admin based
// @ID list-user
// @Accept  json
// @Produce  json
// @Param request body ListUserRequest true "admin list request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /admins/list [get]
func (server *Server) ListAdmins(ctx *gin.Context) {
	var req ListUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.ListUserParams{
		Role:   "admin",
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	userlist, err := server.store.ListUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, userlist)
}

// ListUserStudentRequest contains the impurt parameters for list rolebased user data
type ListUserStudentRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// @Summary ListUserStudent
// @Description ListUserStudent with the provided student based
// @ID list-student
// @Accept  json
// @Produce  json
// @Param request body ListUserStudentRequest true "student list request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /liststudent [get]
func (server *Server) ListUserStudent(ctx *gin.Context) {
	var req ListUserStudentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUserParams{
		Role:   "student",
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	userlist, err := server.store.ListUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, userlist)
}

type GetUserRequest struct {
	UserName string `form:"user_name"`
}

// @Summary Get an user details by username
// @Description Get an user by its username
// @ID get-user
// @Accept json
// @Produce json
// @Param user_name formData string true "UserName"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /user/get [get]
func (server Server) GetUserByUsername(ctx *gin.Context) {
	var req GetUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	// authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	// if req.UserName != authPayload.UserName {
	// 	err := errors.New("account doesn't belong to the authenticated user")
	// 	ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	// 	return
	// }

	user, err := server.store.GetUser(ctx, req.UserName)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

type StudenTCountRequest struct {
	Role string `json:"role"`
}

// Get student Count
func (server *Server) StudentCount(ctx *gin.Context) {
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

type DeleteUserRequest struct {
	UserID int64 `form:"user_id"`
}

// @Summary Delete a user
// @ID delete-user
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /delete/user [delete]
func (server *Server) deleteUsers(ctx *gin.Context) {
	var req DeleteUserRequest
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

	err := server.store.DeleteUsers(ctx, req.UserID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

package api

import (
	db "eduApp/db/sqlc"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VerifyEmailRequest defines the request body structure for creating a submission
type VerifyEmailRequest struct {
	EmailID    int64  `form:"email_id"`
	SecretCode string `form:"secret_code"`
}

type verifyEmailResponse struct {
	IsEmailVerified bool
}

// @Summary Create a new VerifyEmailRequest
// @Description Creates a new Verify Email Request by depend on confirming the email
// @Accept json
// @Produce json
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /verifyemail [post]
// VerifyEmail creates a new course
func (server *Server) VerifyEmail(ctx *gin.Context) {
	var req VerifyEmailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.EmailID,
		SecretCode: req.SecretCode,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := verifyEmailResponse{
		IsEmailVerified: txResult.User.IsEmailVerified,
	}

	ctx.JSON(http.StatusOK, rsp)
}

type GetUserIDRequest struct {
	UserID int64 `json:"user_id"`
}

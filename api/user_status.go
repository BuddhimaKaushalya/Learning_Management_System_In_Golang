package api

import (
	db "eduApp/db/sqlc"
	"eduApp/token"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateUserStatusRequest defines the request body structure for updating a user status
type UpdateUserStatusByAdminRequest struct {
	StatusID int64 `json:"status_id"`
	Active   bool  `json:"active"`
	Pending  bool  `json:"pending"`
}

// @Summary Update an userStatus
// @Description Update the status of a user
// @ID update-userStatus
// @Accept  json
// @Produce  json
// @Param status_id path int true "Status ID"
// @Param request body UpdateUserStatusByAdminRequest true "Update User Status Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /userStatus/edit [put]
// UpdateMaterial updates a use status
func (server *Server) UpdateUserStatusByAdmin(ctx *gin.Context) {
	var req UpdateUserStatusByAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Role != "admin" {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	arg := db.UpdateUserStatusByAdminParams{
		StatusID: req.StatusID,
		Active:   req.Active,
		Pending:  req.Pending,
	}

	userStatus, err := server.store.UpdateUserStatusByAdmin(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, userStatus)
}

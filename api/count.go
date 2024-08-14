package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetCourseCompletedUserCount struct {
	Progress int64 `json:"progress"`
}

// GetCourseCompletedUserCount is a funtion to get count of the course completed use count
func (server *Server) GetCourseCompletedUserCount(ctx *gin.Context) {
	var req GetCourseCompletedUserCount

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	arg := GetCourseCompletedUserCount{
		Progress: int64(100),
	}

	progress, err := server.store.GetCourseCompletedUserCount(ctx, arg.Progress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, progress)
}

// GetTotalCourseCount funtion is api call to get total course count from teh db
func (server *Server) GetTotalCourseCount(ctx *gin.Context) {
	course, err := server.store.GetTotalCourseCount(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "getting count details failed, Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, course)
}

// GetTotalUserCountRequest contains the input parameters of get total User count
type GetTotalUserCountRequest struct {
	Role string `json:"role"`
}

// @Summary Get Total Studet Count
// @Description Count Students using ther role
// @Accept json
// @Produce json
// @Param request body GetTotalUserCountRequest true "Get Total User Count request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /count/student [Get]
// GetTotalStudentCount function is api to call , to get total student count from the in the db
func (server *Server) GetTotalStudentCount(ctx *gin.Context) {
	var req GetTotalUserCountRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := GetTotalUserCountRequest{
		Role: "student",
	}

	user, err := server.store.GetTotalUserCount(ctx, arg.Role)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count user data"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// @Summary Get Total Admin Count
// @Description Count Students using ther role
// @Accept json
// @Produce json
// @Param request body GetTotalUserCountRequest true "Get Total User Count request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /count/admin [Get]
// GetTotalStudentCount function is api to call , to get total admin count from the in the db
func (server *Server) GetTotalAdminCount(ctx *gin.Context) {
	var req GetTotalUserCountRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	// if authPayload.UserName != "admin@123" {
	// 	err := errors.New("authenticated users only able to see these detials, access denied ")
	// 	ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	// }

	arg := GetTotalUserCountRequest{
		Role: "admin",
	}

	user, err := server.store.GetTotalUserCount(ctx, arg.Role)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count admin data"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetTotalSubscribedUserCountRequest contains input parameter for get total subscribed user count
type GetTotalSubscribedUserCountRequest struct {
	Active bool `form:"active"`
}

// @Summary Get Toatal Subscribed user count
// @Description Get Total user Subscription by active status of there corrent status
// @Accept json
// @Produce json
// @Param request body GetTotalSubscribedUserCountRequest true "Get Total Subscribed user count request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /count/subscription [Get]
// GetTotalSubscribedUserCount is a funtion api call to get all course subscribed users
func (server *Server) GetTotalSubscribedUserCount(ctx *gin.Context) {
	var req GetTotalSubscribedUserCountRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := GetTotalSubscribedUserCountRequest{
		Active: true,
	}

	count, err := server.store.GetTotalSubscribedUserCount(ctx, arg.Active)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Get subscribed user count,  Please try agian.!"})
		return
	}

	ctx.JSON(http.StatusOK, count)
}

// Reponse struct contains the response keys for the all count details for the cards
type Response struct {
	CourseCompletedUserCount int64 `json:"course_completed_user_count"`
	TotalSubscribedUserCount int64 `json:"total_subscribed_user_count"`
	TotalStudentUserCount    int64 `json:"total_student_user_count"`
	TotalAdminUserCount      int64 `json:"total_admin_user_count"`
	TotalCourseCount         int64 `json:"total_course_count"`
	InProgressCourseCount    int64 `json:"in_progress_course_count"`
}

// @Summary Count
// @Description  count api able to get all count details at one
// @Accept json
// @Produce json
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /all/counts [Get]
// counts funtion is a api call to get all count details from theh db at once
func (server *Server) counts(ctx *gin.Context) {
	var req1 GetCourseCompletedUserCount
	var req2 GetTotalSubscribedUserCountRequest
	var req3 GetTotalUserCountRequest

	//GetCourseCompletedUserCount
	if err := ctx.ShouldBindQuery(&req1); err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}
	arg := GetCourseCompletedUserCount{
		Progress: int64(100),
	}
	progress, err := server.store.GetCourseCompletedUserCount(ctx, arg.Progress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//GetTotalSubscribedUserCountRequest
	if err := ctx.ShouldBindQuery(&req2); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg1 := GetTotalSubscribedUserCountRequest{
		Active: true,
	}
	SubscribedUserCount, err := server.store.GetTotalSubscribedUserCount(ctx, arg1.Active)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//GetTotalUserCountRequest
	if err := ctx.ShouldBindQuery(&req3); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	TotalStudentUserCount, err := server.store.GetTotalUserCount(ctx, "student")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//GetTotalUserCountRequest
	if err := ctx.ShouldBindQuery(&req3); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	TotalAdminUserCount, err := server.store.GetTotalUserCount(ctx, "admin")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	TotalCourse, err := server.store.GetTotalCourseCount(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	InProgressCourses, err := server.store.GetInProgressCourseCount(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := Response{
		CourseCompletedUserCount: progress,
		TotalSubscribedUserCount: SubscribedUserCount,
		TotalStudentUserCount:    TotalStudentUserCount,
		TotalAdminUserCount:      TotalAdminUserCount,
		TotalCourseCount:         TotalCourse,
		InProgressCourseCount:    InProgressCourses,
	}

	ctx.JSON(http.StatusOK, rsp)
}

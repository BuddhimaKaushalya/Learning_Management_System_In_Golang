package api

import (
	db "eduApp/db/sqlc"
	"eduApp/token"
	"eduApp/util"
	"eduApp/worker"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// server serves hhtp requests
type Server struct {
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	router          *gin.Engine
	taskDistributor worker.TaskDistributor
}

// NewServer creates a http server and setup routing
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("user_name", validUsername)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://testnet.bethelnet.io", "http://*", "https://*", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	grp := router.Group("/static")
	{
		grp.StaticFS("/", http.Dir("./uploads"))
	}
	//	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//public routes
	router.POST("/signup", server.createUser)
	router.POST("/login", server.loginUser)
	router.GET("/verifyemail", server.VerifyEmail)
	router.GET("/getUser/username", server.GetUserByUsername)

	router.GET("/getCount", server.counts)

	router.POST("tokens/renew_access", server.renewAccessToken)

	router.PUT("/lessonCompletion/edit", server.UpdateLessonCompletion)

	//RBAC auth routes
	authroute := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authroute.DELETE("/delete/user", server.deleteUsers)

	authroute.POST("/lessonComplete", server.CreateLessonCompletion)

	//authroute.DELETE("/deleteEmail", server.DeleteEmail)
	authroute.PATCH("/updateUser", server.UpdateUser)
	router.PATCH("/reset/password", server.UpdateUserPassword)
	router.GET("/ckeck/email", server.CheckEmail)

	authroute.POST("/admin/signup", server.createAdminUser)
	router.GET("/user/get", server.GetUserByUsername)

	router.GET("admins/list", server.ListAdmins)

	//Student
	router.GET("/liststudent", server.ListUserStudent)
	router.GET("/student/count", server.StudentCount)
	//courses
	authroute.POST("/course", server.CreateCourse)
	router.GET("/list/course/bycatagory", server.ListAllCourseByCatagory)

	router.GET("/course/get", server.GetCourse)
	router.GET("/courses", server.ListCourses)
	authroute.PATCH("/course/edit", server.UpdateCourse)
	authroute.DELETE("/course/delete", server.DeleteCourse)
	router.GET("/list/catagories", server.ListAllCourseCatagories)

	//Assignment
	authroute.POST("/assignments", server.CreateAssignment)
	authroute.GET("/assignment/get", server.getAssignment)
	authroute.PATCH("/assignments/edit", server.UpdateAssignment)
	authroute.DELETE("/assignment/delete", server.deleteAssignment)

	//category
	authroute.POST("/category", server.CreateCategory)
	authroute.GET("/category/get", server.GetCategory)
	authroute.PATCH("/category/edit", server.UpdateCategory)
	authroute.DELETE("/category/delete", server.DeleteCategory)
	router.GET("/categories", server.ListAllCategories)

	//submissions
	authroute.POST("/submission/create", server.CreateSubmission)
	authroute.GET("/submission/byassignment", server.GetSubmissionsByAssignment)
	authroute.GET("/submission/byuser", server.GetSubmissionsByUser)
	authroute.GET("/submissions", server.listSubmissions)
	authroute.PUT("/submission/edit", server.UpdateSubmission)
	authroute.DELETE("/submission/delete", server.DeleteSubmission)

	//Materials
	authroute.POST("/material", server.CreateMaterial)
	authroute.GET("/material/get", server.GetMaterial)
	authroute.GET("/materialCount/get", server.GetTotalMaterialInCourse)
	authroute.PUT("/material/edit", server.UpdateMaterial)
	authroute.DELETE("/material/delete", server.DeleteMaterial)
	authroute.GET("/material/list", server.ListMaterial)

	//course progress
	authroute.GET("/courseprogress", server.ListCourseProgressByUser)
	authroute.GET("/courseprogress/get", server.getCourseProgress)
	authroute.DELETE("/courseprogress/delete", server.DeleteCourseProgress)

	//user Status
	authroute.PUT("/userStatus/edit", server.UpdateUserStatusByAdmin)
	// authroute.DELETE("/userStatus/delete", server.DeleteUserStatus)

	// Student Marks
	authroute.POST("/create/mark", server.CreateMark)
	authroute.DELETE("/mark/delete", server.DeleteMark)
	authroute.GET("/mark/get", server.GetMark)
	authroute.GET("/marks/bycourse", server.ListMarks)
	authroute.PUT("/mark/edit", server.UpdateMark)

	// Progress
	router.POST("/createprogress", server.createCourseProgress)
	authroute.GET("/progress/get", server.getCourseProgress)
	authroute.GET("/progress/list", server.ListCourseProgressByUser)
	authroute.DELETE("/courseProgress/delete ", server.DeleteCourseProgress)

	//router.PUT("/progress/edit", server.UpdateCourseProgress)

	//Subscription
	authroute.POST("/subscription", server.CreateSubscription)
	authroute.GET("/subscription/get", server.GetSubscription)
	authroute.GET("/subscription/user", server.ListSubscriptionsByUser)
	authroute.GET("/subscription/course", server.ListSubscriptionsByCourse)
	authroute.PUT("/subscription/edit", server.UpdateSubscriptions)
	authroute.GET("/count/course/subscription", server.GetUserCountForCertianCourse)

	//Request
	router.POST("/request/create", server.CreateRequest)
	authroute.PUT("/request/edit", server.UpdateRequest)

	//Profile Picture
	authroute.POST("/profile", server.CreateProfilePicture)
	authroute.PATCH("/profile/edit", server.UpdateProfilePicture)
	authroute.DELETE("/profile/delete", server.DeleteProfilePicture)

	//count
	router.GET("/count/student", server.GetTotalStudentCount)
	router.GET("/count/admin", server.GetTotalAdminCount)
	router.GET("/count/subscription", server.GetTotalSubscribedUserCount)
	router.GET("/all/counts", server.counts)

	server.router = router
}

// start runs the http server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

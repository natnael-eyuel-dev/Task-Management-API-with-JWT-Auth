package router

//imports
import (
	"github.com/gin-gonic/gin"
	"github.com/natnael-eyuel-dev/Task-Management-API-with-JWT-Auth/controllers"
	"github.com/natnael-eyuel-dev/Task-Management-API-with-JWT-Auth/data"
	"github.com/natnael-eyuel-dev/Task-Management-API-with-JWT-Auth/middleware"
)

func SetupRouter(taskService data.TaskManager, userService data.UserService) *gin.Engine {
	router := gin.Default()     // create default gin router

	taskController := controllers.NewTaskController(taskService)      // inject task service into task controller
	userConroller := controllers.NewUserController(userService)       // inject user service into user controller

	// authenticated routes 
	authMiddleWare := middleware.AuthMiddleWare()
	
	authGroup := router.Group("/")
	authGroup.Use(authMiddleWare)
	{
		authGroup.GET("/tasks", taskController.GetAllTasks)          // get all tasks
		authGroup.GET("/tasks/:id", taskController.GetTaskByID)      // get specific task by id
	}

	// admin only routes 
	adminOnly := middleware.AdminOnly()

	adminGroup := router.Group("/")
	adminGroup.Use(authMiddleWare, adminOnly)
	{
		adminGroup.POST("/tasks", taskController.CreateTask)            // create new task
		adminGroup.DELETE("/tasks/:id", taskController.DeleteTask)      // delete task by id
		adminGroup.PUT("/tasks/:id", taskController.UpdateTask)         // update existing task
		adminGroup.PUT("/promote/:id", userConroller.PromoteAdmin)      // promote user to admin
	}
	
	// public routes
	router.POST("/register", userConroller.Register)        // register new user
	router.POST("/login", userConroller.Login)              // authenticate a user

	return router     // return configured router
} 
package main

import (
	"log"
	"task-manager/backend/internal/handlers"
	"task-manager/backend/internal/middleware"
	"task-manager/backend/internal/repositories"
	"task-manager/backend/internal/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	dbCfg := repositories.NewDatabaseConfig()
	db, err := dbCfg.Connect()
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance: ", err)
	}
	defer sqlDB.Close()

	registerService := services.NewRegisterService()
	registrationHandler := handlers.NewRegisterHandler(db, registerService)

	authService := services.NewAuthService()
	authHandler := handlers.NewAuthHandler(db, authService)

	taskService := services.NewTaskService()
	taskHandler := handlers.NewTaskHandler(db, taskService)

	refreshHandler := handlers.NewRefreshHandler(db, authService)

	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(db, userService)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://host.docker.internal"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	v1 := r.Group("/api/v1")
	{
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", registrationHandler.Registration)
			authRoutes.POST("/login", authHandler.Token)
			authRoutes.POST("/refresh", refreshHandler.Refresh)
		}

		// Task routes with ABAC policies
		taskRoutes := v1.Group("/tasks")
		taskRoutes.Use(middleware.AuthMiddleware())
		{
			// Create task - any authenticated user with task:create permission
			taskRoutes.POST("", middleware.RequirePermission("tasks", "create"), taskHandler.CreateTask)

			// Update task - user must own the task or be admin with task:update permission
			taskRoutes.PUT("/:id", middleware.RequirePermission("tasks", "update"), taskHandler.UpdateTask)

			// Delete task - user must own the task or be admin with task:delete permission
			taskRoutes.DELETE("/:id", middleware.RequirePermission("tasks", "delete"), taskHandler.DeleteTask)

			// Get specific task - user must own the task or be admin with task:read permission
			taskRoutes.GET("/:id", middleware.RequirePermission("tasks", "read"), taskHandler.GetTaskByID)

			// Get all tasks - admin only with task:read permission
			taskRoutes.GET("", middleware.RequireRoleAndPermission("admin", "tasks", "read"), taskHandler.GetTasks)
		}

		// User routes with ABAC policies
		userRoutes := v1.Group("/users")
		userRoutes.Use(middleware.AuthMiddleware())
		{
			// Delete user - admin only with user:delete permission
			userRoutes.DELETE("/:user_id", middleware.RequireRoleAndPermission("admin", "users", "delete"), userHandler.DeleteUser)

			// Get all users - admin only with user:read permission
			userRoutes.GET("", middleware.RequireRoleAndPermission("admin", "users", "read"), userHandler.GetUsers)

			// Get tasks by user - admin can access any user's tasks, regular users can only access their own
			userRoutes.GET("/:user_id/tasks", middleware.RequirePermission("tasks", "read"), taskHandler.GetTasksByUser)

			// Get own profile - any authenticated user
			userRoutes.GET("/profile", userHandler.GetUserProfile)

			// Get user profile by ID - admin only with user:read permission
			userRoutes.GET("/profile/:user_id", middleware.RequireRoleAndPermission("admin", "users", "read"), userHandler.GetUserProfileByUserId)
		}

		// Admin-only routes
		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
		{
			adminRoutes.GET("/dashboard", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "admin access granted"})
			})
		}
	}
	r.Run(":8080")
}

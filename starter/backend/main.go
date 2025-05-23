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

	userHandler := handlers.NewUserHandler(db, nil)

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
		taskRoutes := v1.Group("/tasks")
		taskRoutes.Use(middleware.AuthMiddleware())
		{
			taskRoutes.POST("", taskHandler.CreateTask)
			taskRoutes.PUT("/:id", taskHandler.UpdateTask)
			taskRoutes.DELETE("/:id", taskHandler.DeleteTask)
			taskRoutes.GET("/:id", taskHandler.GetTaskByID)
			taskRoutes.GET("", taskHandler.GetTasks)
		}
		userRoutes := v1.Group("/users")
		{
			userRoutes.DELETE("/:user_id", userHandler.DeleteUser)
			userRoutes.GET("", userHandler.GetUsers)
			userRoutes.GET("/:user_id/tasks", taskHandler.GetTasksByUser)
			userRoutes.GET("/profile", userHandler.GetUserProfile)
			userRoutes.GET("/profile/:user_id", userHandler.GetUserProfileByUserId)
		}

		// Admin-only route
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

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"task-manager/backend/internal/handlers"
	"task-manager/backend/internal/middleware"
	"task-manager/backend/internal/models"
	"task-manager/backend/internal/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupABACTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.RolePermission{}, &models.UserRole{}, &models.Token{}, &models.Task{})
	assert.NoError(t, err)

	// Create default roles
	adminRole := models.Role{
		ID:   uuid.Must(uuid.NewV4()),
		Name: "admin",
	}
	userRole := models.Role{
		ID:   uuid.Must(uuid.NewV4()),
		Name: "user",
	}
	db.Create(&adminRole)
	db.Create(&userRole)

	// Create permissions
	permissions := []models.Permission{
		{ID: uuid.Must(uuid.NewV4()), Resource: "tasks", Action: "create"},
		{ID: uuid.Must(uuid.NewV4()), Resource: "tasks", Action: "read"},
		{ID: uuid.Must(uuid.NewV4()), Resource: "tasks", Action: "update"},
		{ID: uuid.Must(uuid.NewV4()), Resource: "tasks", Action: "delete"},
		{ID: uuid.Must(uuid.NewV4()), Resource: "users", Action: "create"},
		{ID: uuid.Must(uuid.NewV4()), Resource: "users", Action: "read"},
		{ID: uuid.Must(uuid.NewV4()), Resource: "users", Action: "update"},
		{ID: uuid.Must(uuid.NewV4()), Resource: "users", Action: "delete"},
	}
	for _, perm := range permissions {
		db.Create(&perm)
	}

	// Assign permissions to roles
	for _, perm := range permissions {
		// Admin gets all permissions
		db.Create(&models.RolePermission{
			RoleID:       adminRole.ID,
			PermissionID: perm.ID,
		})

		// User gets only task permissions
		if perm.Resource == "tasks" {
			db.Create(&models.RolePermission{
				RoleID:       userRole.ID,
				PermissionID: perm.ID,
			})
		}
	}

	return db
}

func createTestUser(t *testing.T, db *gorm.DB, username, email, password string, isAdmin bool) (uuid.UUID, string) {
	// Create user
	user := models.User{
		ID:       uuid.Must(uuid.NewV4()),
		Username: username,
		Email:    email,
		Password: password,
	}
	db.Create(&user)

	// Assign role
	var role models.Role
	if isAdmin {
		db.Where("name = ?", "admin").First(&role)
	} else {
		db.Where("name = ?", "user").First(&role)
	}

	db.Create(&models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	})

	// Generate token
	authService := services.NewAuthService()
	accessToken, _, err := authService.GenerateToken(db, user.ID, user.Username)
	assert.NoError(t, err)

	return user.ID, accessToken
}

func TestABACTaskAccessPolicies(t *testing.T) {
	db := setupABACTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup handlers
	taskService := services.NewTaskService()
	taskHandler := handlers.NewTaskHandler(db, taskService)
	authService := services.NewAuthService()
	authHandler := handlers.NewAuthHandler(db, authService)
	registerHandler := handlers.NewRegisterHandler(db, services.NewRegisterService())

	// Setup routes with ABAC policies
	router.POST("/register", registerHandler.Registration)
	router.POST("/login", authHandler.Token)

	// Task routes with ABAC policies
	taskRoutes := router.Group("/tasks")
	taskRoutes.Use(middleware.AuthMiddleware())
	{
		taskRoutes.POST("", middleware.RequirePermission("tasks", "create"), taskHandler.CreateTask)
		taskRoutes.PUT("/:id", middleware.RequirePermission("tasks", "update"), taskHandler.UpdateTask)
		taskRoutes.DELETE("/:id", middleware.RequirePermission("tasks", "delete"), taskHandler.DeleteTask)
		taskRoutes.GET("/:id", middleware.RequirePermission("tasks", "read"), taskHandler.GetTaskByID)
		taskRoutes.GET("", middleware.RequireRoleAndPermission("admin", "tasks", "read"), taskHandler.GetTasks)
	}

	// Create test users
	_, adminToken := createTestUser(t, db, "admin", "admin@test.com", "admin123", true)
	user1ID, user1Token := createTestUser(t, db, "user1", "user1@test.com", "user123", false)
	user2ID, _ := createTestUser(t, db, "user2", "user2@test.com", "user123", false)

	// Create test tasks
	task1 := models.Task{
		ID:          uuid.Must(uuid.NewV4()),
		Title:       "Task 1",
		Description: "Task 1 description",
		UserID:      user1ID,
		Status:      "pending",
	}
	task2 := models.Task{
		ID:          uuid.Must(uuid.NewV4()),
		Title:       "Task 2",
		Description: "Task 2 description",
		UserID:      user2ID,
		Status:      "pending",
	}
	db.Create(&task1)
	db.Create(&task2)

	t.Run("Admin can access all tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Regular user cannot access all tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("User can access their own task", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/"+task1.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("User cannot access another user's task", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/"+task2.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("Admin can access any user's task", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/"+task1.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("User can update their own task", func(t *testing.T) {
		updateData := map[string]interface{}{
			"title": "Updated Task 1",
		}
		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/tasks/"+task1.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+user1Token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("User cannot update another user's task", func(t *testing.T) {
		updateData := map[string]interface{}{
			"title": "Updated Task 2",
		}
		body, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/tasks/"+task2.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+user1Token)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestABACUserAccessPolicies(t *testing.T) {
	db := setupABACTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup handlers
	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(db, userService)
	authService := services.NewAuthService()
	authHandler := handlers.NewAuthHandler(db, authService)

	// Setup routes with ABAC policies
	router.POST("/login", authHandler.Token)

	userRoutes := router.Group("/users")
	userRoutes.Use(middleware.AuthMiddleware())
	{
		userRoutes.DELETE("/:user_id", middleware.RequireRoleAndPermission("admin", "users", "delete"), userHandler.DeleteUser)
		userRoutes.GET("", middleware.RequireRoleAndPermission("admin", "users", "read"), userHandler.GetUsers)
		userRoutes.GET("/profile", userHandler.GetUserProfile)
		userRoutes.GET("/profile/:user_id", middleware.RequireRoleAndPermission("admin", "users", "read"), userHandler.GetUserProfileByUserId)
	}

	// Create test users
	adminID, adminToken := createTestUser(t, db, "admin", "admin@test.com", "admin123", true)
	user1ID, user1Token := createTestUser(t, db, "user1", "user1@test.com", "user123", false)

	t.Run("Admin can get all users", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Regular user cannot get all users", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users", nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("Admin can get any user's profile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/profile/"+user1ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Regular user cannot get another user's profile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/profile/"+adminID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("User can get their own profile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Admin can delete users", func(t *testing.T) {
		// Create a user to delete
		deleteUserID, _ := createTestUser(t, db, "deleteuser", "delete@test.com", "delete123", false)

		req := httptest.NewRequest("DELETE", "/users/"+deleteUserID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("Regular user cannot delete users", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/"+adminID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

func TestABACPermissionEnforcement(t *testing.T) {
	db := setupABACTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup middleware test routes
	router.GET("/test-permission/:resource/:action", middleware.AuthMiddleware(), func(c *gin.Context) {
		resource := c.Param("resource")
		action := c.Param("action")
		middleware.RequirePermission(resource, action)(c)
		if !c.IsAborted() {
			c.JSON(http.StatusOK, gin.H{"message": "permission granted"})
		}
	})

	router.GET("/test-role-permission/:role/:resource/:action", middleware.AuthMiddleware(), func(c *gin.Context) {
		role := c.Param("role")
		resource := c.Param("resource")
		action := c.Param("action")
		middleware.RequireRoleAndPermission(role, resource, action)(c)
		if !c.IsAborted() {
			c.JSON(http.StatusOK, gin.H{"message": "role and permission granted"})
		}
	})

	// Create test users
	_, adminToken := createTestUser(t, db, "admin", "admin@test.com", "admin123", true)
	_, user1Token := createTestUser(t, db, "user1", "user1@test.com", "user123", false)

	t.Run("Admin has all permissions", func(t *testing.T) {
		permissions := []string{"tasks:create", "tasks:read", "tasks:update", "tasks:delete", "users:read", "users:delete"}
		for _, perm := range permissions {
			parts := strings.Split(perm, ":")
			resource := parts[0]
			action := parts[1]
			req := httptest.NewRequest("GET", "/test-permission/"+resource+"/"+action, nil)
			req.Header.Set("Authorization", "Bearer "+adminToken)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusOK, resp.Code, "Admin should have permission: %s", perm)
		}
	})

	t.Run("User has limited permissions", func(t *testing.T) {
		// User should have task permissions
		taskPermissions := []string{"tasks:create", "tasks:read", "tasks:update", "tasks:delete"}
		for _, perm := range taskPermissions {
			parts := strings.Split(perm, ":")
			resource := parts[0]
			action := parts[1]
			req := httptest.NewRequest("GET", "/test-permission/"+resource+"/"+action, nil)
			req.Header.Set("Authorization", "Bearer "+user1Token)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusOK, resp.Code, "User should have task permission: %s", perm)
		}

		// User should not have user management permissions
		userPermissions := []string{"users:read", "users:delete"}
		for _, perm := range userPermissions {
			parts := strings.Split(perm, ":")
			resource := parts[0]
			action := parts[1]
			req := httptest.NewRequest("GET", "/test-permission/"+resource+"/"+action, nil)
			req.Header.Set("Authorization", "Bearer "+user1Token)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusForbidden, resp.Code, "User should not have user permission: %s", perm)
		}
	})

	t.Run("Role and permission combination works", func(t *testing.T) {
		// Admin should have admin role with any permission
		req := httptest.NewRequest("GET", "/test-role-permission/admin/users/read", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		// User should not have admin role
		req = httptest.NewRequest("GET", "/test-role-permission/admin/users/read", nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

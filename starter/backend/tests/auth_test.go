package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"task-manager/backend/internal/handlers"
	"task-manager/backend/internal/middleware"
	"task-manager/backend/internal/models"
	"task-manager/backend/internal/services"
	"task-manager/backend/internal/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.RolePermission{}, &models.UserRole{}, &models.Token{})
	assert.NoError(t, err)

	// Create default roles and permissions
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
		{
			ID:       uuid.Must(uuid.NewV4()),
			Resource: "tasks",
			Action:   "create",
		},
		{
			ID:       uuid.Must(uuid.NewV4()),
			Resource: "tasks",
			Action:   "read",
		},
		{
			ID:       uuid.Must(uuid.NewV4()),
			Resource: "tasks",
			Action:   "update",
		},
		{
			ID:       uuid.Must(uuid.NewV4()),
			Resource: "tasks",
			Action:   "delete",
		},
	}
	for _, perm := range permissions {
		db.Create(&perm)
	}

	// Assign permissions to roles
	for _, perm := range permissions {
		db.Create(&models.RolePermission{
			RoleID:       adminRole.ID,
			PermissionID: perm.ID,
		})
		if perm.Action == "read" {
			db.Create(&models.RolePermission{
				RoleID:       userRole.ID,
				PermissionID: perm.ID,
			})
		}
	}

	return db
}

func TestRegistrationWithRole(t *testing.T) {
	db := setupTestDB(t)
	registerService := services.NewRegisterService()

	// Test user registration
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := registerService.RegisterUser(db, user)
	assert.NoError(t, err)

	// Verify user was created with user role
	var createdUser models.User
	err = db.Preload("Roles").Where("username = ?", "testuser").First(&createdUser).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, len(createdUser.Roles))
	assert.Equal(t, "user", createdUser.Roles[0].Name)
}

func TestLoginWithRolesAndPermissions(t *testing.T) {
	db := setupTestDB(t)
	authService := services.NewAuthService()

	// Create a test user
	registerService := services.NewRegisterService()
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	err := registerService.RegisterUser(db, user)
	assert.NoError(t, err)

	// Test login
	loggedInUser, err := authService.LoginUser(db, "testuser", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, loggedInUser)

	// Generate token
	accessToken, refreshToken, err := authService.GenerateToken(db, loggedInUser.ID, loggedInUser.Username)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// Verify token contains roles and permissions
	claims, err := utils.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.Contains(t, claims.Roles, "user")
	assert.Contains(t, claims.Permissions, "tasks:read")
}

func TestProtectedRoutes(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes with middleware
	authService := services.NewAuthService()
	authHandler := handlers.NewAuthHandler(db, authService)
	registerHandler := handlers.NewRegisterHandler(db, services.NewRegisterService())

	router.POST("/register", registerHandler.Registration)
	router.POST("/login", authHandler.Token)

	// Protected route that requires admin role
	router.GET("/admin-only", middleware.AuthMiddleware(), middleware.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Protected route that requires specific permission
	router.GET("/tasks", middleware.AuthMiddleware(), middleware.RequirePermission("tasks", "read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "tasks access granted"})
	})

	// Test registration
	registerReq := handlers.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerHTTPReq := httptest.NewRequest("POST", "/register", bytes.NewBuffer(registerBody))
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerHTTPReq)
	assert.Equal(t, http.StatusCreated, registerResp.Code)

	// Test login
	loginReq := handlers.AuthRequest{
		Username: "testuser",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginHTTPReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(loginBody))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginHTTPReq)
	assert.Equal(t, http.StatusOK, loginResp.Code)

	var loginResponse handlers.AuthResponse
	json.Unmarshal(loginResp.Body.Bytes(), &loginResponse)

	// Test accessing admin-only route (should fail)
	adminReq := httptest.NewRequest("GET", "/admin-only", nil)
	adminReq.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
	adminResp := httptest.NewRecorder()
	router.ServeHTTP(adminResp, adminReq)
	assert.Equal(t, http.StatusForbidden, adminResp.Code)

	// Test accessing tasks route (should succeed)
	tasksReq := httptest.NewRequest("GET", "/tasks", nil)
	tasksReq.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
	tasksResp := httptest.NewRecorder()
	router.ServeHTTP(tasksResp, tasksReq)
	assert.Equal(t, http.StatusOK, tasksResp.Code)
}

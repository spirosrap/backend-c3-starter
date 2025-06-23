package handlers

import (
	"net/http"
	"task-manager/backend/internal/models"
	"task-manager/backend/internal/services"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type TaskHandler struct {
	db          *gorm.DB
	taskService services.TaskService
}

func NewTaskHandler(db *gorm.DB, taskService services.TaskService) *TaskHandler {
	return &TaskHandler{db: db, taskService: taskService}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from the JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var userUUID uuid.UUID
	switch v := userID.(type) {
	case uuid.UUID:
		userUUID = v
	case string:
		u, err := uuid.FromString(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}
		userUUID = u
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id type"})
		return
	}

	// Set the user ID from the token
	task.UserID = userUUID

	// Debug logging
	log.Printf("[DEBUG] userUUID: %v (type: %T)", userUUID, userUUID)
	log.Printf("[DEBUG] task: %+v", task)

	// Generate a new UUID for the task if not set
	if task.ID == uuid.Nil {
		newID, err := uuid.NewV4()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate task ID"})
			return
		}
		task.ID = newID
	}

	if err := h.taskService.CreateTask(h.db, &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	taskID, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// Get user info from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID := userID.(uuid.UUID)

	// Check if user is admin
	userRoles, exists := c.Get("roles")
	isAdmin := false
	if exists {
		rolesList := userRoles.([]string)
		for _, role := range rolesList {
			if role == "admin" {
				isAdmin = true
				break
			}
		}
	}

	task, err := h.taskService.GetTaskByID(h.db, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// Enforce ownership: only task owner or admin can access
	if !isAdmin && task.UserID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied - you can only access your own tasks"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	// This endpoint is already protected by RequireRoleAndPermission("admin", "tasks", "read")
	// So only admins can access all tasks
	tasks, err := h.taskService.GetTasks(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	taskID, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// Get user info from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID := userID.(uuid.UUID)

	// Check if user is admin
	userRoles, exists := c.Get("roles")
	isAdmin := false
	if exists {
		rolesList := userRoles.([]string)
		for _, role := range rolesList {
			if role == "admin" {
				isAdmin = true
				break
			}
		}
	}

	// Get the existing task to check ownership
	existingTask, err := h.taskService.GetTaskByID(h.db, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// Enforce ownership: only task owner or admin can update
	if !isAdmin && existingTask.UserID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied - you can only update your own tasks"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the task belongs to the same user (unless admin)
	if !isAdmin {
		task.UserID = userUUID
	}

	if err := h.taskService.UpdateTask(h.db, taskID, &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task updated successfully"})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	taskID, err := uuid.FromString(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// Get user info from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userUUID := userID.(uuid.UUID)

	// Check if user is admin
	userRoles, exists := c.Get("roles")
	isAdmin := false
	if exists {
		rolesList := userRoles.([]string)
		for _, role := range rolesList {
			if role == "admin" {
				isAdmin = true
				break
			}
		}
	}

	// Get the existing task to check ownership
	existingTask, err := h.taskService.GetTaskByID(h.db, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// Enforce ownership: only task owner or admin can delete
	if !isAdmin && existingTask.UserID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied - you can only delete your own tasks"})
		return
	}

	if err := h.taskService.DeleteTask(h.db, taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *TaskHandler) GetTasksByUser(c *gin.Context) {
	userID := c.Param("user_id")
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get current user info from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	currentUserUUID := currentUserID.(uuid.UUID)

	// Check if current user is admin
	userRoles, exists := c.Get("roles")
	isAdmin := false
	if exists {
		rolesList := userRoles.([]string)
		for _, role := range rolesList {
			if role == "admin" {
				isAdmin = true
				break
			}
		}
	}

	// Enforce access control: users can only access their own tasks, admins can access any user's tasks
	if !isAdmin && currentUserUUID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied - you can only access your own tasks"})
		return
	}

	var tasks []models.Task
	if err := h.db.Where("user_id = ?", userUUID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

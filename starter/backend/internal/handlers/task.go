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
	task, err := h.taskService.GetTaskByID(h.db, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
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
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	var tasks []models.Task
	if err := h.db.Where("user_id = ?", userUUID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

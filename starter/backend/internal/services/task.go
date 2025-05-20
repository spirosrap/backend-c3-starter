package services

import (
	"errors"
	"task-manager/backend/internal/models"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type TaskService interface {
	CreateTask(db *gorm.DB, task *models.Task) error
	GetTaskByID(db *gorm.DB, taskID uuid.UUID) (*models.Task, error)
	GetTasks(db *gorm.DB) ([]models.Task, error)
	UpdateTask(db *gorm.DB, taskID uuid.UUID, task *models.Task) error
	DeleteTask(db *gorm.DB, taskID uuid.UUID) error
}

type TaskServiceImpl struct{}

func NewTaskService() *TaskServiceImpl {
	return &TaskServiceImpl{}
}

func (s *TaskServiceImpl) CreateTask(db *gorm.DB, task *models.Task) error {
	return db.Create(task).Error
}

func (s *TaskServiceImpl) GetTaskByID(db *gorm.DB, taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	if err := db.First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskServiceImpl) GetTasks(db *gorm.DB) ([]models.Task, error) {
	var tasks []models.Task
	if err := db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskServiceImpl) UpdateTask(db *gorm.DB, taskID uuid.UUID, task *models.Task) error {
	result := db.Model(&models.Task{}).Where("id = ?", taskID).Updates(task)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found")
	}
	return nil
}

func (s *TaskServiceImpl) DeleteTask(db *gorm.DB, taskID uuid.UUID) error {
	result := db.Delete(&models.Task{}, "id = ?", taskID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found")
	}
	return nil
}

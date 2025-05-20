package services

import (
	"task-manager/backend/internal/models"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterService interface {
	RegisterUser(db *gorm.DB, user models.User) error
}

type RegisterServiceImpl struct{}

func NewRegisterService() *RegisterServiceImpl {
	return &RegisterServiceImpl{}
}

func (s *RegisterServiceImpl) RegisterUser(db *gorm.DB, user models.User) error {
	// Generate a new UUID for the user
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	user.ID = id

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Get the default user role
	var userRole models.Role
	if err := db.Where("name = ?", "user").First(&userRole).Error; err != nil {
		return err
	}

	// Assign the default user role
	user.Roles = []models.Role{userRole}

	// Create the user in the database
	result := db.Create(&user)
	return result.Error
}

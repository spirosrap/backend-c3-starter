package services

import (
	"errors"
	"log"
	"task-manager/backend/internal/models"
	"task-manager/backend/internal/utils"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	LoginUser(db *gorm.DB, username, password string) (*models.User, error)
	GenerateToken(db *gorm.DB, userID uuid.UUID, username string) (string, string, error)
}

type AuthServiceImpl struct {
}

func NewAuthService() *AuthServiceImpl {
	return &AuthServiceImpl{}
}

func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (s *AuthServiceImpl) LoginUser(db *gorm.DB, username, password string) (*models.User, error) {
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		log.Printf("Database error during login: %v", err)
		return nil, errors.New("internal server error")
	}

	if !VerifyPassword(user.Password, password) {
		return nil, errors.New("invalid username or password")
	}

	return &user, nil
}

func (s *AuthServiceImpl) GenerateToken(db *gorm.DB, userID uuid.UUID, username string) (string, string, error) {
	// Generate access token
	accessToken, err := utils.GenerateAccessToken(userID, username)
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		return "", "", errors.New("failed to generate access token")
	}

	// Generate refresh token
	refreshToken, err := uuid.NewV4()
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		return "", "", errors.New("failed to generate refresh token")
	}

	// Store refresh token in database
	token := models.Token{
		ID:           uuid.Must(uuid.NewV4()),
		UserId:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	if err := db.Create(&token).Error; err != nil {
		log.Printf("Error storing refresh token: %v", err)
		return "", "", errors.New("failed to store refresh token")
	}

	return accessToken, refreshToken.String(), nil
}

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
	RefreshToken(db *gorm.DB, refreshToken string) (string, string, error)
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

func (s *AuthServiceImpl) RefreshToken(db *gorm.DB, refreshToken string) (string, string, error) {
	// Parse the refresh token
	tokenUUID, err := uuid.FromString(refreshToken)
	if err != nil {
		log.Printf("Invalid refresh token format: %v", err)
		return "", "", errors.New("invalid refresh token")
	}

	// Find the token in the database
	var token models.Token
	if err := db.Where("refresh_token = ?", tokenUUID).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid refresh token")
		}
		log.Printf("Database error during token refresh: %v", err)
		return "", "", errors.New("internal server error")
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	// Get the user
	var user models.User
	if err := db.First(&user, token.UserId).Error; err != nil {
		log.Printf("Error finding user: %v", err)
		return "", "", errors.New("internal server error")
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := s.GenerateToken(db, user.ID, user.Username)
	if err != nil {
		return "", "", err
	}

	// Delete the old token
	if err := db.Delete(&token).Error; err != nil {
		log.Printf("Error deleting old token: %v", err)
		// Continue anyway as we've already generated new tokens
	}

	return accessToken, newRefreshToken, nil
}

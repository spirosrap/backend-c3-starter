package models

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	ID          uuid.UUID    `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"unique"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

type Permission struct {
	gorm.Model
	ID       uuid.UUID `json:"id" gorm:"primaryKey"`
	Resource string    `json:"resource"`
	Action   string    `json:"action"`
	Roles    []Role    `json:"roles" gorm:"many2many:role_permissions;"`
}

type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" gorm:"primaryKey"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"primaryKey"`
}

type UserRole struct {
	UserID uuid.UUID `json:"user_id" gorm:"primaryKey"`
	RoleID uuid.UUID `json:"role_id" gorm:"primaryKey"`
}

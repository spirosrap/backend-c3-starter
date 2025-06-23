package middleware

import (
	"net/http"
	"strings"
	"task-manager/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateAccessToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// RequireRole checks if the user has any of the required roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user roles not found"})
			c.Abort()
			return
		}

		rolesList := userRoles.([]string)
		for _, requiredRole := range roles {
			for _, userRole := range rolesList {
				if requiredRole == userRole {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions - role required"})
		c.Abort()
	}
}

// RequirePermission checks if the user has the required permission
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user permissions not found"})
			c.Abort()
			return
		}

		permissionsList := permissions.([]string)
		requiredPermission := resource + ":" + action

		for _, permission := range permissionsList {
			if permission == requiredPermission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions - permission required"})
		c.Abort()
	}
}

// RequireRoleAndPermission combines role and permission checks
func RequireRoleAndPermission(role string, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check role first
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user roles not found"})
			c.Abort()
			return
		}

		rolesList := userRoles.([]string)
		hasRole := false
		for _, userRole := range rolesList {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions - role required"})
			c.Abort()
			return
		}

		// Check permission
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user permissions not found"})
			c.Abort()
			return
		}

		permissionsList := permissions.([]string)
		requiredPermission := resource + ":" + action

		for _, permission := range permissionsList {
			if permission == requiredPermission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions - permission required"})
		c.Abort()
	}
}

// RequireOwnershipOrAdmin checks if the user owns the resource or is an admin
func RequireOwnershipOrAdmin(resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		// Check if user is admin
		userRoles, exists := c.Get("roles")
		if exists {
			rolesList := userRoles.([]string)
			for _, role := range rolesList {
				if role == "admin" {
					c.Next()
					return
				}
			}
		}

		// Check ownership
		resourceID := c.Param(resourceIDParam)
		if resourceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource ID not provided"})
			c.Abort()
			return
		}

		// Parse resource ID
		resourceUUID, err := uuid.FromString(resourceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource ID format"})
			c.Abort()
			return
		}

		// Check if the resource belongs to the user
		userUUID := userID.(uuid.UUID)
		if resourceUUID == userUUID {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "access denied - resource ownership required"})
		c.Abort()
	}
}

// ABAC Policies for specific endpoints

// RequireTaskAccess checks if user can access a specific task
func RequireTaskAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is admin
		userRoles, exists := c.Get("roles")
		if exists {
			rolesList := userRoles.([]string)
			for _, role := range rolesList {
				if role == "admin" {
					c.Next()
					return
				}
			}
		}

		// For non-admin users, they can only access their own tasks
		// This will be enforced at the service layer
		c.Next()
	}
}

// RequireUserManagementAccess checks if user can manage other users
func RequireUserManagementAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only admins can manage users
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user roles not found"})
			c.Abort()
			return
		}

		rolesList := userRoles.([]string)
		for _, role := range rolesList {
			if role == "admin" {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions - admin role required for user management"})
		c.Abort()
	}
}

// RequireTaskManagementAccess checks if user can manage tasks
func RequireTaskManagementAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user roles not found"})
			c.Abort()
			return
		}

		// Admin can manage all tasks
		rolesList := userRoles.([]string)
		for _, role := range rolesList {
			if role == "admin" {
				c.Next()
				return
			}
		}

		// Regular users can manage their own tasks
		// This will be enforced at the service layer
		c.Next()
	}
}

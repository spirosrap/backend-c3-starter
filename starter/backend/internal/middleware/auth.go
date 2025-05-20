package middleware

import (
	"net/http"
	"strings"
	"task-manager/backend/internal/utils"

	"github.com/gin-gonic/gin"
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

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

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

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

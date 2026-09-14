package middleware

import (
	"net/http"
	"strings"

	"findit-backend/utils"
	"github.com/gin-gonic/gin"
)

const (
	UserIDKey   = "userID"
	UserEmailKey = "userEmail"
	UserRoleKey  = "userRole"
)

// AuthMiddleware validates JWT Bearer tokens for protected endpoints
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Akses ditolak: Token otentikasi tidak ditemukan")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Akses ditolak: Format Authorization header harus 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims, err := utils.ValidateToken(tokenStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Akses ditolak: Token otentikasi tidak valid atau sudah kadaluarsa")
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UserRoleKey, claims.Role)

		c.Next()
	}
}

// GetAuthenticatedUserID extracts the user ID set by AuthMiddleware
func GetAuthenticatedUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := val.(uint)
	return id, ok
}

// GetAuthenticatedUserRole extracts the user role set by AuthMiddleware
func GetAuthenticatedUserRole(c *gin.Context) (string, bool) {
	val, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}

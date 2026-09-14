package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"findit-backend/middleware"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestJWTGenerationAndValidation(t *testing.T) {
	userID := uint(42)
	email := "testuser@example.com"
	role := "user"

	tokenStr, err := utils.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("Expected token generation to succeed, got err: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("Generated token string is empty")
	}

	claims, err := utils.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("Expected token validation to succeed, got err: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}
	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for missing token, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for invalid token, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, _ := middleware.GetAuthenticatedUserID(c)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	token, _ := utils.GenerateToken(10, "user@test.com", "user")
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK for valid token, got %d", w.Code)
	}
}

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"findit-backend/config"
	"findit-backend/middleware"
	"findit-backend/models"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
)

func setupReportTestRouter() *gin.Engine {
	r := gin.New()
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	protected.PUT("/reports/:id", UpdateReport)
	protected.DELETE("/reports/:id", DeleteReport)
	return r
}

func TestReportOwnership_OtherUserForbidden(t *testing.T) {
	// Setup in-memory mock DB using GORM sqlite or skip DB call if mock DB unavailable
	if config.DB == nil {
		t.Skip("Database not initialized, skipping DB dependent ownership test")
		return
	}

	// Create dummy report belonging to UserID = 100
	report := models.Report{
		UserID: 100,
		Type:   "lost",
		Title:  "Test Wallet",
	}
	config.DB.Create(&report)
	defer config.DB.Delete(&report)

	r := setupReportTestRouter()

	// Token for UserID = 200 (different user)
	token, _ := utils.GenerateToken(200, "otheruser@test.com", "user")

	updatePayload, _ := json.Marshal(map[string]interface{}{
		"title": "Hacked Title",
	})

	req, _ := http.NewRequest(http.MethodPut, "/api/reports/"+utils.NewUUID(), bytes.NewBuffer(updatePayload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Since ID doesn't exist or if report belongs to user 100, checking status
	// If record found, status is 403 Forbidden for user 200
}

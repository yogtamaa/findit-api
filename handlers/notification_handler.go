package handlers

import (
	"errors"
	"net/http"
	"time"

	"findit-backend/config"
	"findit-backend/models"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateNotification creates a new notification.
// Since notifications.id is a UUID string (not auto-increment), we
// generate it here before inserting.
func CreateNotification(c *gin.Context) {
	var notification models.Notification

	if err := c.ShouldBindJSON(&notification); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: "+err.Error())
		return
	}
	notification.ID = utils.NewUUID()

	if err := config.DB.Create(&notification).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan notifikasi: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Notifikasi berhasil dibuat", notification)
}

// GetNotificationsForUser retrieves notifications for a given user
// (notifiable_type = "User", notifiable_id = the given user id).
func GetNotificationsForUser(c *gin.Context) {
	userID := c.Param("userId")
	var notifications []models.Notification

	if err := config.DB.
		Where("notifiable_type = ? AND notifiable_id = ?", "User", userID).
		Order("created_at desc").
		Find(&notifications).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil notifikasi")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil notifikasi", notifications)
}

// MarkNotificationRead sets read_at to the current time for a notification.
func MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	var notification models.Notification

	if err := config.DB.First(&notification, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Notifikasi tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	now := time.Now()
	if err := config.DB.Model(&notification).Update("read_at", &now).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui notifikasi")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notifikasi ditandai sudah dibaca", notification)
}

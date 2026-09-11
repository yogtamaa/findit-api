package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"findit-backend/config"
	"findit-backend/models"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetReports retrieves reports, newest first.
// Supports optional query filters: ?type=lost|found  &status=baru|dicocokkan|...
func GetReports(c *gin.Context) {
	var reports []models.Report
	query := config.DB.Order("created_at desc")

	if t := c.Query("type"); t != "" {
		query = query.Where("type = ?", t)
	}
	if s := c.Query("status"); s != "" {
		query = query.Where("status = ?", s)
	}

	if err := query.Find(&reports).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data laporan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil daftar laporan", reports)
}

// GetReportByID retrieves a single report by ID.
func GetReportByID(c *gin.Context) {
	id := c.Param("id")
	var report models.Report

	if err := config.DB.First(&report, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Laporan tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil detail laporan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil detail laporan", report)
}

// CreateReport handles creation of a new lost/found report.
func CreateReport(c *gin.Context) {
	var report models.Report

	if err := c.ShouldBindJSON(&report); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: "+err.Error())
		return
	}

	if report.Status == "" {
		report.Status = "baru"
	}
	if report.ReportIdentifier == "" {
		prefix := "LST"
		if report.Type == "found" {
			prefix = "FND"
		}
		report.ReportIdentifier = fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}

	if err := config.DB.Create(&report).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan laporan: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Laporan berhasil dibuat", report)
}

// UpdateReport handles updating an existing report by ID.
func UpdateReport(c *gin.Context) {
	id := c.Param("id")
	var report models.Report

	if err := config.DB.First(&report, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Laporan tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	var input models.Report
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data input tidak valid: "+err.Error())
		return
	}

	if err := config.DB.Model(&report).Updates(map[string]interface{}{
		"title":         input.Title,
		"description":   input.Description,
		"category":      input.Category,
		"location":      input.Location,
		"photo_url":     input.PhotoURL,
		"status":        input.Status,
		"activity_note": input.ActivityNote,
		"item_date":     input.ItemDate,
	}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui laporan")
		return
	}

	config.DB.First(&report, id)
	utils.SuccessResponse(c, http.StatusOK, "Berhasil memperbarui laporan", report)
}

// DeleteReport deletes a report by ID.
func DeleteReport(c *gin.Context) {
	id := c.Param("id")
	var report models.Report

	if err := config.DB.First(&report, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Laporan tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	if err := config.DB.Delete(&report).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus laporan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan berhasil dihapus", nil)
}

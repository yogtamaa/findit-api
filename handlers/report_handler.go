package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"findit-backend/config"
	"findit-backend/middleware"
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

	// Normalisasi status kosong -> "baru" sebagai defensive fallback untuk
	// data legacy/nyasar (mis. row yang tersimpan sebelum perbaikan header,
	// atau client lama yang mengirim status ""). Akar masalah sebenarnya bukan
	// schema DB: kolom `status` sudah VARCHAR(30) DEFAULT 'baru' dan handler
	// create sudah set "baru" sebelum INSERT. Kode ini tetap dipertahankan
	// supaya respons konsisten walau ada row dengan status kosong di DB.
	for i := range reports {
		if reports[i].Status == "" {
			reports[i].Status = "baru"
		}
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
	if report.Status == "" {
		report.Status = "baru"
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

	// Auto fill UserID from JWT token if available
	if authID, ok := middleware.GetAuthenticatedUserID(c); ok {
		report.UserID = authID
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

// UpdateReport handles updating an existing report by ID with ownership verification.
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

	// Ownership check
	if authID, ok := middleware.GetAuthenticatedUserID(c); ok {
		role, _ := middleware.GetAuthenticatedUserRole(c)
		if report.UserID != authID && role != "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Akses ditolak: Anda tidak memiliki wewenang untuk mengubah laporan ini")
			return
		}
	}

	// Parse sebagai map (bukan models.Report) supaya field yang TIDAK dikirim
	// tidak ikut ditimpa jadi kosong. Cocok untuk update sebagian (misal cuma "status").
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data input tidak valid: "+err.Error())
		return
	}

	allowedFields := []string{"title", "description", "category", "room_number", "location", "photo_url", "status", "activity_note", "item_date"}
	updates := map[string]interface{}{}
	for _, field := range allowedFields {
		if val, ok := raw[field]; ok {
			updates[field] = val
		}
	}

	if len(updates) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tidak ada field valid untuk diperbarui")
		return
	}

	if err := config.DB.Model(&report).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui laporan")
		return
	}

	config.DB.First(&report, id)
	if report.Status == "" {
		report.Status = "baru"
	}
	utils.SuccessResponse(c, http.StatusOK, "Berhasil memperbarui laporan", report)
}

// DeleteReport deletes a report by ID with ownership verification.
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

	// Ownership check
	if authID, ok := middleware.GetAuthenticatedUserID(c); ok {
		role, _ := middleware.GetAuthenticatedUserRole(c)
		if report.UserID != authID && role != "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Akses ditolak: Anda tidak memiliki wewenang untuk menghapus laporan ini")
			return
		}
	}

	if err := config.DB.Delete(&report).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus laporan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan berhasil dihapus", nil)
}

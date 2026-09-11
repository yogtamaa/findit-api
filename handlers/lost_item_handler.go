package handlers

import (
	"errors"
	"net/http"

	"findit-backend/config"
	"findit-backend/models"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetLostItems retrieves all lost item reports from MySQL
func GetLostItems(c *gin.Context) {
	var items []models.LostItem

	if err := config.DB.Order("created_at desc").Find(&items).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data barang hilang")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil daftar barang hilang", items)
}

// GetLostItemByID retrieves a single lost item report by ID
func GetLostItemByID(c *gin.Context) {
	id := c.Param("id")
	var item models.LostItem

	if err := config.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Data barang hilang tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil detail barang hilang")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil detail barang hilang", item)
}

// CreateLostItem handles creation of a new lost item report
func CreateLostItem(c *gin.Context) {
	var item models.LostItem

	if err := c.ShouldBindJSON(&item); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: judul barang wajib diisi")
		return
	}

	if item.Status == "" {
		item.Status = "LOST"
	}

	if err := config.DB.Create(&item).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan laporan barang hilang")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Laporan barang hilang berhasil dibuat", item)
}

// UpdateLostItem handles updating an existing lost item report by ID
func UpdateLostItem(c *gin.Context) {
	id := c.Param("id")
	var item models.LostItem

	if err := config.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Data barang hilang tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	var input models.LostItem
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data input tidak valid")
		return
	}

	// Update fields
	if err := config.DB.Model(&item).Updates(map[string]interface{}{
		"user_id":     input.UserID,
		"title":       input.Title,
		"description": input.Description,
		"category":    input.Category,
		"location":    input.Location,
		"contact":     input.Contact,
		"status":      input.Status,
		"date":        input.Date,
	}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui data barang hilang")
		return
	}

	config.DB.First(&item, id)
	utils.SuccessResponse(c, http.StatusOK, "Berhasil memperbarui data barang hilang", item)
}

// DeleteLostItem deletes a lost item report by ID
func DeleteLostItem(c *gin.Context) {
	id := c.Param("id")
	var item models.LostItem

	if err := config.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Data barang hilang tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	if err := config.DB.Delete(&item).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus data barang hilang")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan barang hilang berhasil dihapus", nil)
}

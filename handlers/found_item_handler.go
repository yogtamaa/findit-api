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

// GetFoundItems retrieves all found item reports from MySQL
func GetFoundItems(c *gin.Context) {
	var items []models.FoundItem

	if err := config.DB.Order("created_at desc").Find(&items).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data barang ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil daftar barang ditemukan", items)
}

// GetFoundItemByID retrieves a single found item report by ID
func GetFoundItemByID(c *gin.Context) {
	id := c.Param("id")
	var item models.FoundItem

	if err := config.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Data barang ditemukan tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil detail barang ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil detail barang ditemukan", item)
}

// CreateFoundItem handles creation of a new found item report
func CreateFoundItem(c *gin.Context) {
	var item models.FoundItem

	if err := c.ShouldBindJSON(&item); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: judul barang wajib diisi")
		return
	}

	if item.Status == "" {
		item.Status = "FOUND"
	}

	if err := config.DB.Create(&item).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan laporan barang ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Laporan barang ditemukan berhasil dibuat", item)
}

// UpdateFoundItem handles updating an existing found item report by ID
func UpdateFoundItem(c *gin.Context) {
	id := c.Param("id")
	var item models.FoundItem

	if err := config.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Data barang ditemukan tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	var input models.FoundItem
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui data barang ditemukan")
		return
	}

	config.DB.First(&item, id)
	utils.SuccessResponse(c, http.StatusOK, "Berhasil memperbarui data barang ditemukan", item)
}

// DeleteFoundItem deletes a found item report by ID
func DeleteFoundItem(c *gin.Context) {
	id := c.Param("id")
	var item models.FoundItem

	if err := config.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Data barang ditemukan tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	if err := config.DB.Delete(&item).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus data barang ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan barang ditemukan berhasil dihapus", nil)
}

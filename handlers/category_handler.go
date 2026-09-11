package handlers

import (
	"net/http"

	"findit-backend/config"
	"findit-backend/models"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
)

// GetCategories retrieves all categories, used to populate dropdowns in the app.
func GetCategories(c *gin.Context) {
	var categories []models.Category

	if err := config.DB.Order("name asc").Find(&categories).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kategori")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil daftar kategori", categories)
}

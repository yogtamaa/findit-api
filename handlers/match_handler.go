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

// GetMatches retrieves all matches, newest first.
func GetMatches(c *gin.Context) {
	var matches []models.Match

	if err := config.DB.Order("created_at desc").Find(&matches).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data matches")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil daftar matches", matches)
}

// GetMatchByID retrieves a single match by ID.
func GetMatchByID(c *gin.Context) {
	id := c.Param("id")
	var match models.Match

	if err := config.DB.First(&match, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Match tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil detail match")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil detail match", match)
}

// CreateMatch creates a new match linking a lost report with a found report.
func CreateMatch(c *gin.Context) {
	var match models.Match

	if err := c.ShouldBindJSON(&match); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: "+err.Error())
		return
	}
	if match.Status == "" {
		match.Status = "pending"
	}

	if err := config.DB.Create(&match).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan match: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Match berhasil dibuat", match)
}

// UpdateMatch updates a match's status/handover info (e.g. verified, handover method, contact shared timestamp).
func UpdateMatch(c *gin.Context) {
	id := c.Param("id")
	var match models.Match

	if err := config.DB.First(&match, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Match tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	var input models.Match
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data input tidak valid: "+err.Error())
		return
	}

	if err := config.DB.Model(&match).Updates(map[string]interface{}{
		"status":            input.Status,
		"verified_by":       input.VerifiedBy,
		"handover_method":   input.HandoverMethod,
		"contact_shared_at": input.ContactSharedAt,
		"activity_note":     input.ActivityNote,
	}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui match")
		return
	}

	config.DB.First(&match, id)
	utils.SuccessResponse(c, http.StatusOK, "Berhasil memperbarui match", match)
}

package handlers

import (
	"errors"
	"net/http"
	"path/filepath"

	"findit-backend/storage"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
)

// UploadImage handles secure image upload requests (protected by AuthMiddleware)
func UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		// Try alternative form keys: "image" or "photo"
		fileHeader, err = c.FormFile("image")
		if err != nil {
			fileHeader, err = c.FormFile("photo")
		}
	}

	if err != nil || fileHeader == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File upload tidak ditemukan, harap sertakan field 'file'")
		return
	}

	provider := storage.NewLocalStorageProvider()
	fileURL, err := provider.SaveFile(fileHeader)
	if err != nil {
		if errors.Is(err, storage.ErrFileTooLarge) {
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, err.Error())
			return
		}
		if errors.Is(err, storage.ErrInvalidFileType) || errors.Is(err, storage.ErrPathTraversal) {
			utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memproses upload file: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Foto berhasil diunggah", gin.H{
		"url":      fileURL,
		"filename": filepath.Base(fileURL),
		"size":     fileHeader.Size,
	})
}

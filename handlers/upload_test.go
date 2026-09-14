package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"findit-backend/middleware"
	"findit-backend/storage"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
)

// 1x1 pixel PNG bytes
var validPNGBytes = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
	0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

func setupUploadRouter() *gin.Engine {
	r := gin.New()
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/upload", UploadImage)
	return r
}

func TestUpload_WithoutJWT(t *testing.T) {
	r := setupUploadRouter()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.png")
	part.Write(validPNGBytes)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for missing JWT, got %d", w.Code)
	}
}

func TestUpload_InvalidFileType(t *testing.T) {
	r := setupUploadRouter()
	token, _ := utils.GenerateToken(1, "user@test.com", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "script.sh")
	part.Write([]byte("#!/bin/bash\necho 'malicious'"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request for non-image script file, got %d", w.Code)
	}
}

func TestUpload_FileTooLarge(t *testing.T) {
	r := setupUploadRouter()
	token, _ := utils.GenerateToken(1, "user@test.com", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.png")
	part.Write(make([]byte, 5*1024*1024+1))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413 Request Entity Too Large, got %d", w.Code)
	}
}

func TestUpload_ValidImage(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("UPLOAD_DIR", tempDir)
	defer os.Unsetenv("UPLOAD_DIR")

	r := setupUploadRouter()
	token, _ := utils.GenerateToken(1, "user@test.com", "user")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "sample.png")
	part.Write(validPNGBytes)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created for valid image upload, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestStorage_PathTraversalPrevention(t *testing.T) {
	tempDir := t.TempDir()
	provider := &storage.LocalStorageProvider{
		UploadDir:    tempDir,
		BaseURL:      "/uploads",
		MaxSizeBytes: 5 * 1024 * 1024,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "../../etc/passwd.png")
	part.Write(validPNGBytes)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_ = req.ParseMultipartForm(5 * 1024 * 1024)

	fileHeader := req.MultipartForm.File["file"][0]

	url, err := provider.SaveFile(fileHeader)
	if err != nil {
		t.Fatalf("SaveFile failed unexpectedly: %v", err)
	}

	// Verify filename was sanitized to UUID and stored inside tempDir only
	filename := filepath.Base(url)
	targetPath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Expected file to be safely saved under upload dir, but not found at %s", targetPath)
	}
}

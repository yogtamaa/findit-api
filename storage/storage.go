package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"findit-backend/utils"
)

var (
	ErrFileTooLarge    = errors.New("ukuran file melebihi batas maksimum (5 MB)")
	ErrInvalidFileType = errors.New("tipe file tidak didukung, hanya format JPEG, PNG, dan WebP yang diizinkan")
	ErrPathTraversal   = errors.New("nama file mengandung karakter tidak aman / path traversal")
)

// StorageProvider defines the contract for uploading and deleting files
type StorageProvider interface {
	SaveFile(fileHeader *multipart.FileHeader) (string, error)
	DeleteFile(fileURLOrPath string) error
}

// LocalStorageProvider implements StorageProvider for local filesystem storage
type LocalStorageProvider struct {
	UploadDir string
	BaseURL   string
	MaxSizeBytes int64
}

// NewLocalStorageProvider initializes LocalStorageProvider using environment variables
func NewLocalStorageProvider() *LocalStorageProvider {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	baseURL := os.Getenv("STORAGE_BASE_URL")
	if baseURL == "" {
		baseURL = "/uploads"
	}

	maxMB := int64(5)
	if maxMBEnv := os.Getenv("MAX_UPLOAD_SIZE_MB"); maxMBEnv != "" {
		if val, err := strconv.ParseInt(maxMBEnv, 10, 64); err == nil && val > 0 {
			maxMB = val
		}
	}

	return &LocalStorageProvider{
		UploadDir:    uploadDir,
		BaseURL:      strings.TrimRight(baseURL, "/"),
		MaxSizeBytes: maxMB * 1024 * 1024,
	}
}

// SaveFile validates and saves an uploaded file securely
func (p *LocalStorageProvider) SaveFile(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", errors.New("file upload tidak boleh kosong")
	}

	// 1. Enforce file size limit
	if fileHeader.Size > p.MaxSizeBytes {
		return "", ErrFileTooLarge
	}

	// 2. Open file stream
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file upload: %w", err)
	}
	defer src.Close()

	// 3. Read initial 512 bytes to detect MIME type via magic bytes
	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("gagal membaca header file: %w", err)
	}

	contentType := http.DetectContentType(buf[:n])
	allowedMimeTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	ext, isAllowed := allowedMimeTypes[contentType]
	if !isAllowed {
		return "", ErrInvalidFileType
	}

	// 4. Reset read pointer after reading header bytes
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("gagal memproses file upload: %w", err)
	}

	// 5. Sanitize original filename and prevent path traversal
	cleanOriginalName := filepath.Base(fileHeader.Filename)
	if strings.Contains(cleanOriginalName, "..") || strings.ContainsAny(cleanOriginalName, "/\\") {
		return "", ErrPathTraversal
	}

	// 6. Generate clean UUID filename
	newFilename := fmt.Sprintf("%s%s", utils.NewUUID(), ext)

	// 7. Ensure upload directory exists
	if err := os.MkdirAll(p.UploadDir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori penyimpan: %w", err)
	}

	// 8. Create target file path safely
	targetPath := filepath.Join(p.UploadDir, newFilename)

	// Extra path traversal sanity check
	absUploadDir, _ := filepath.Abs(p.UploadDir)
	absTargetPath, _ := filepath.Abs(targetPath)
	if !strings.HasPrefix(absTargetPath, absUploadDir) {
		return "", ErrPathTraversal
	}

	dst, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file ke disk: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("gagal menulis konten file: %w", err)
	}

	// 9. Construct returned public URL path
	publicURL := fmt.Sprintf("%s/%s", p.BaseURL, newFilename)
	return publicURL, nil
}

// DeleteFile deletes an uploaded file from disk if it exists
func (p *LocalStorageProvider) DeleteFile(fileURLOrPath string) error {
	if fileURLOrPath == "" {
		return nil
	}

	filename := filepath.Base(fileURLOrPath)
	if filename == "" || filename == "." || filename == "/" {
		return nil
	}

	targetPath := filepath.Join(p.UploadDir, filename)
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

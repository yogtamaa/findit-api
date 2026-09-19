package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"findit-backend/config"
	"findit-backend/middleware"
	"findit-backend/models"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type registerInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	// Role opsional. Reguler hanya boleh membuat pekerja/user — admin dilarang.
	Role string `json:"role"`
}

type loginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,max=72"`
}

type updateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Password string `json:"password" binding:"omitempty,min=6,max=72"`
}

// normalizeRole memetakan role input ke nilai baku. Role admin tidak boleh
// dibuat lewat endpoint publik maupun dikelola oleh non-admin.
func normalizeRole(role string) (string, error) {
	switch role {
	case "", "user":
		return "user", nil
	case "worker":
		return "worker", nil
	case "admin":
		return "", errors.New("role 'admin' tidak dapat dibuat via endpoint ini")
	default:
		return "", errors.New("role tidak dikenal: " + role)
	}
}

// requireAdmin memeriksa role JWT aktif. Kembalikan false bila bukan admin
// (respons 403 sudah ditulis).
func requireAdmin(c *gin.Context) bool {
	role, ok := middleware.GetAuthenticatedUserRole(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Akses ditolak: sesi tidak terautentikasi")
		return false
	}
	if role != "admin" {
		utils.ErrorResponse(c, http.StatusForbidden, "Akses ditolak: hanya administrator yang dapat mengelola pekerja")
		return false
	}
	return true
}

func parseUserID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID user tidak valid")
		return 0, false
	}
	return uint(id), true
}

// Register creates a new user with a securely hashed password.
func Register(c *gin.Context) {
	var input registerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: "+err.Error())
		return
	}

	role, err := normalizeRole(input.Role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: "+err.Error())
		return
	}

	var existing models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email sudah terdaftar")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memproses password")
		return
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Phone:    input.Phone,
		Role:     role,
		Password: string(hashed),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat akun: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Akun berhasil dibuat", user)
}

// Login verifies email + password and returns the user profile with a JWT.
func Login(c *gin.Context) {
	var input loginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: "+err.Error())
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Email atau password salah")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan pada database")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat token otentikasi")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login berhasil", gin.H{
		"token": token,
		"user":  user,
	})
}

// GetUserByID retrieves a single user profile (password is never included, see models.User json tags).
func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := config.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data user")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data user", user)
}

// GetWorkers lists all non-admin staff (room attendants / workers) ordered by newest first.
func GetWorkers(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	var workers []models.User
	if err := config.DB.Where("role <> ?", "admin").Order("id DESC").Find(&workers).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data pekerja")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data pekerja", workers)
}

// UpdateUser updates name/email/phone/role and optionally the password (re-hashed).
func UpdateUser(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	id, ok := parseUserID(c)
	if !ok {
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data user")
		return
	}

	var input updateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: "+err.Error())
		return
	}

	updates := map[string]interface{}{}

	if input.Name != "" {
		updates["name"] = input.Name
	}

	if input.Email != "" {
		if input.Email != user.Email {
			var existing models.User
			if err := config.DB.Where("email = ? AND id <> ?", input.Email, user.ID).First(&existing).Error; err == nil {
				utils.ErrorResponse(c, http.StatusConflict, "Email sudah digunakan user lain")
				return
			}
		}
		updates["email"] = input.Email
	}

	if input.Phone != "" {
		updates["phone"] = input.Phone
	}

	if input.Role != "" {
		if user.Role == "admin" && input.Role != "admin" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Role akun administrator tidak dapat diubah")
			return
		}
		normalized, err := normalizeRole(input.Role)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Validasi gagal: "+err.Error())
			return
		}
		updates["role"] = normalized
	}

	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memproses password")
			return
		}
		updates["password"] = string(hashed)
	}

	if len(updates) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tidak ada field yang diubah")
		return
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui user: "+err.Error())
		return
	}

	config.DB.First(&user, id)
	utils.SuccessResponse(c, http.StatusOK, "User berhasil diperbarui", user)
}

// DeleteUser removes a non-admin worker. Akun admin dan akun sendiri dilindungi.
func DeleteUser(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	authID, ok := middleware.GetAuthenticatedUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Akses ditolak: sesi tidak terautentikasi")
		return
	}

	id, ok := parseUserID(c)
	if !ok {
		return
	}

	if id == authID {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tidak dapat menghapus akun yang sedang dipakai")
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data user")
		return
	}

	if user.Role == "admin" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Akun administrator tidak dapat dihapus")
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusConflict, "Gagal menghapus: pekerja masih memiliki data terkait (laporan/aktivitas)")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pekerja berhasil dihapus", nil)
}
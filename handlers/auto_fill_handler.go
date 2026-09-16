package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"findit-backend/config"
	"findit-backend/models"
	"findit-backend/storage"
	"findit-backend/utils"
	"github.com/gin-gonic/gin"
)

// confidenceScore menyimpan confidence sebagai number (0.0 - 1.0) untuk response
// API, tetapi mampu menerima dua format hasil dari Gemini: string
// "high"|"medium"|"low" (sesuai prompt) atau number (0.0 - 1.0).
type confidenceScore float64

func (c *confidenceScore) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*c = 0
		return nil
	}

	if s[0] == '"' {
		var raw string
		if err := json.Unmarshal(b, &raw); err != nil {
			return err
		}
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "high":
			*c = 1.0
			return nil
		case "medium":
			*c = 0.5
			return nil
		case "low":
			*c = 0.2
			return nil
		}
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			*c = confidenceScore(f)
			return nil
		}
		*c = 0
		return nil
	}

	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	*c = confidenceScore(f)
	return nil
}

// geminiAnalysis adalah hasil parse dari response Gemini Vision (field title,
// description, category, confidence sesuai prompt yang dikirimkan).
type geminiAnalysis struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Confidence  confidenceScore `json:"confidence"`
}

// Struktur request/response Gemini generateContent (hanya bagian yang dibutuhkan).
type geminiFile struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiPart struct {
	Text       string      `json:"text,omitempty"`
	Thought    bool        `json:"thought,omitempty"`
	InlineData *geminiFile `json:"inline_data,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiJSONSchema struct {
	Type       string                      `json:"type"`
	Properties map[string]geminiSchemaProp `json:"properties"`
	Required   []string                    `json:"required,omitempty"`
}

type geminiSchemaProp struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type geminiThinkingConfig struct {
	ThinkingLevel string `json:"thinkingLevel,omitempty"`
}

type geminiGenerationConfig struct {
	ResponseMIMEType string                `json:"responseMimeType,omitempty"`
	ResponseSchema   *geminiJSONSchema     `json:"responseSchema,omitempty"`
	ThinkingConfig   *geminiThinkingConfig `json:"thinkingConfig,omitempty"`
	Temperature      float64               `json:"temperature,omitempty"`
	MaxOutputTokens  int                   `json:"maxOutputTokens,omitempty"`
}

type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type geminiCandidate struct {
	Content      geminiResponseContent `json:"content"`
	FinishReason string                `json:"finishReason"`
}

type geminiResponseContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

// AutoFillFromPhoto menerima foto barang, mengunggahnya ke storage, lalu meminta
// Gemini Vision untuk memberikan saran title, description, category, dan confidence.
// Jika Gemini gagal/timeout, foto tetap terunggah dan field lain dikembalikan kosong
// supaya Room Attendant masih bisa mengisi manual (fitur tidak pernah blocking).
func AutoFillFromPhoto(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		// Fallback ke key alternatif: "image" atau "photo"
		fileHeader, err = c.FormFile("image")
		if err != nil {
			fileHeader, err = c.FormFile("photo")
		}
	}

	if err != nil || fileHeader == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File upload tidak ditemukan, harap sertakan field 'file'")
		return
	}

	// Upload foto dulu — pola yang sama dengan POST /upload
	provider := storage.NewLocalStorageProvider()
	photoURL, err := provider.SaveFile(fileHeader)
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

	// Kerangka response: photo_url SELALU terisi jika upload berhasil.
	// Field lain baru diisi setelah analisis AI berhasil.
	result := gin.H{
		"photo_url":   photoURL,
		"title":       "",
		"description": "",
		"category":    "",
		"confidence":  0,
	}

	// Ambil daftar kategori dari database untuk dijadikan opsi di prompt Gemini.
	var categories []models.Category
	if err := config.DB.Order("name asc").Find(&categories).Error; err != nil {
		log.Printf("⚠️ [auto-fill] Gagal memuat kategori: %v", err)
		utils.SuccessResponse(c, http.StatusOK, "Foto berhasil diunggah, namun gagal memuat kategori untuk analisis", result)
		return
	}

	categoryNames := make([]string, 0, len(categories))
	for _, cat := range categories {
		categoryNames = append(categoryNames, cat.Name)
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("⚠️ [auto-fill] Gagal membuka file: %v", err)
		utils.SuccessResponse(c, http.StatusOK, "Foto berhasil diunggah, namun gagal membaca file untuk analisis", result)
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("⚠️ [auto-fill] Gagal membaca isi file: %v", err)
		utils.SuccessResponse(c, http.StatusOK, "Foto berhasil diunggah, namun gagal membaca file untuk analisis", result)
		return
	}

	mimeType := http.DetectContentType(imageBytes)

	analysis, err := analyzeImageWithGemini(imageBytes, mimeType, categoryNames)
	if err != nil {
		// Vision API gagal/timeout -> tetap return 200 dengan photo_url,
		// field lain kosong supaya RA isi manual.
		log.Printf("⚠️ [auto-fill] Vision API gagal: %v", err)
		utils.SuccessResponse(c, http.StatusOK, "Foto berhasil diunggah, namun analisis AI gagal. Silakan isi manual.", result)
		return
	}

	result["title"] = analysis.Title
	result["description"] = analysis.Description
	result["category"] = analysis.Category
	result["confidence"] = analysis.Confidence

	utils.SuccessResponse(c, http.StatusOK, "Analisis foto berhasil", result)
}

// analyzeImageWithGemini memanggil Gemini generateContent dengan foto (base64) dan
// daftar kategori, lalu mengembalikan hasil analisis. Mengembalikan error jika API
// key tidak ada, request timeout, response error, atau JSON tidak bisa diparse.
func analyzeImageWithGemini(imageBytes []byte, mimeType string, categoryNames []string) (*geminiAnalysis, error) {
	apiKey := config.GetGeminiAPIKey()
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY belum di-set di environment")
	}

	model := config.GetGeminiModel()
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		model, apiKey,
	)

	categoriesText := strings.Join(categoryNames, ", ")
	if categoriesText == "" {
		categoriesText = "(tidak ada kategori terdaftar)"
	}

	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	genConfig := &geminiGenerationConfig{
		Temperature:     0.2,
		MaxOutputTokens: 1024,
	}
	// Model 3.x punya thinking bawaan yang bikin respon lama (terutama foto
	// gelap/kurang jelas). Set thinkingLevel minimal supaya lebih cepat.
	// Matikan via GEMINI_THINKING_LEVEL=off, sesuaikan level via env.
	if strings.Contains(model, "3.") {
		level := strings.TrimSpace(os.Getenv("GEMINI_THINKING_LEVEL"))
		if level == "" {
			level = "minimal"
		}
		if l := strings.ToLower(level); l != "off" && l != "false" {
			genConfig.ThinkingConfig = &geminiThinkingConfig{ThinkingLevel: level}
		}
	}
	// JSON mode: minta Gemini balikin JSON persis bentuk responseSchema.
	// Gemini-2.0/2.5/3.x mendukung ini. Matikan via GEMINI_JSON_MODE=false
	// kalau pakai model yang tidak support structured output.
	if strings.ToLower(strings.TrimSpace(os.Getenv("GEMINI_JSON_MODE"))) != "false" {
		genConfig.ResponseMIMEType = "application/json"
		genConfig.ResponseSchema = &geminiJSONSchema{
			Type: "object",
			Properties: map[string]geminiSchemaProp{
				"title":       {Type: "string", Description: "Judul singkat barang dalam bahasa Indonesia"},
				"description": {Type: "string", Description: "Deskripsi singkat barang (warna, merek, kondisi) dalam bahasa Indonesia"},
				"category":    {Type: "string", Description: "Kategori barang (dari daftar kategori yang diberikan, atau kategori singkat baru)"},
				"confidence":  {Type: "string", Description: "Tingkat keyakinan identifikasi: high/medium/low", Enum: []string{"high", "medium", "low"}},
			},
			Required: []string{"title", "description", "category", "confidence"},
		}
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{InlineData: &geminiFile{MIMEType: mimeType, Data: base64Image}},
					{Text: buildGeminiPrompt(categoriesText)},
				},
			},
		},
		GenerationConfig: genConfig,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	timeout := 60 * time.Second
	if envTimeout := os.Getenv("GEMINI_TIMEOUT_SECONDS"); envTimeout != "" {
		if secs, err := strconv.Atoi(envTimeout); err == nil && secs > 0 {
			timeout = time.Duration(secs) * time.Second
		}
	}
	client := &http.Client{Timeout: timeout}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request ke Gemini gagal: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("response Gemini tidak valid: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini error %d: %s", geminiResp.Error.Code, geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("Gemini tidak mengembalikan hasil")
	}

	candidate := geminiResp.Candidates[0]

	// Gabungkan semua text part (skip part thought). Di beberapa model jawaban
	// asli tidak selalu di parts[0], jadi jangan cuma baca bagian pertama.
	var textBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Thought {
			continue
		}
		if strings.TrimSpace(part.Text) != "" {
			textBuilder.WriteString(part.Text)
		}
	}
	rawText := textBuilder.String()

	if strings.TrimSpace(rawText) == "" {
		return nil, fmt.Errorf("Gemini mengembalikan konten kosong (finishReason=%s)", candidate.FinishReason)
	}

	text := extractJSON(rawText)

	var analysis geminiAnalysis
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		// Log text mentah supaya mudah di-debug kalau parsing gagal.
		snippet := rawText
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		log.Printf("⚠️ [auto-fill] Gagal parse JSON dari Gemini (finishReason=%s). Text mentah: %q", candidate.FinishReason, snippet)
		return nil, fmt.Errorf("gagal parse JSON dari Gemini: %w (finishReason=%s)", err, candidate.FinishReason)
	}

	return &analysis, nil
}

// buildGeminiPrompt menyusun prompt analisis foto. Bisa di-override lewat env
// GEMINI_AUTOFILL_PROMPT; gunakan placeholder {{categories}} untuk daftar kategori.
func buildGeminiPrompt(categoriesText string) string {
	if custom := strings.TrimSpace(os.Getenv("GEMINI_AUTOFILL_PROMPT")); custom != "" {
		return strings.ReplaceAll(custom, "{{categories}}", categoriesText)
	}

	return fmt.Sprintf(`Kamu adalah asisten untuk aplikasi "lost & found" hotel bernama FindIt.
Seorang Room Attendant menemukan barang di kamar hotel dan mengambil foto barang tersebut.
Analisis foto ini dan identifikasi barang yang terlihat pada foto.

Kategori yang tersedia (pilih persis salah satu jika cocok; jika tidak ada yang cocok, buat kategori singkat dalam bahasa Indonesia):
%s

Berikan hasil HANYA dalam format JSON berikut (tanpa teks lain):
{
  "title": "judul singkat barang dalam bahasa Indonesia",
  "description": "deskripsi singkat (warna, merek, kondisi, dll) dalam bahasa Indonesia",
  "category": "kategori dari daftar di atas (atau kategori baru singkat)",
  "confidence": 0.0
}

confidence adalah angka 0 sampai 1 yang menunjukkan seberapa yakin kamu terhadap hasil identifikasi.`, categoriesText)
}

// extractJSON membersihkan text response Gemini: membuang markdown code fence (```json)
// dan memotong hanya bagian objek JSON pertama.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}

	if start := strings.Index(s, "{"); start != -1 {
		if end := strings.LastIndex(s, "}"); end != -1 && end > start {
			s = s[start : end+1]
		}
	}

	return s
}

package config

import (
	"os"
	"strings"
)

// ServiceConfig holds service-related configuration values
type ServiceConfig struct {
	GoAPIBaseURL string
	GeminiAPIKey string
	GeminiModel  string
}

// LoadServiceConfig loads service configurations from environment variables
func LoadServiceConfig() ServiceConfig {
	baseURL := os.Getenv("GO_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8094/api"
	}

	return ServiceConfig{
		GoAPIBaseURL: baseURL,
		GeminiAPIKey: GetGeminiAPIKey(),
		GeminiModel:  GetGeminiModel(),
	}
}

// GetGeminiAPIKey retrieves the GEMINI_API_KEY environment variable.
func GetGeminiAPIKey() string {
	return strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
}

// GetGeminiModel retrieves the GEMINI_MODEL environment variable with a default fallback.
func GetGeminiModel() string {
	model := os.Getenv("GEMINI_MODEL")
	if strings.TrimSpace(model) == "" {
		return "gemini-3.6-flash"
	}
	return strings.TrimSpace(model)
}

// GetGoAPIBaseURL retrieves the GO_API_BASE_URL environment variable with a default fallback
func GetGoAPIBaseURL() string {
	baseURL := os.Getenv("GO_API_BASE_URL")
	if baseURL == "" {
		return "http://localhost:8094/api"
	}
	return baseURL
}

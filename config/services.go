package config

import "os"

// ServiceConfig holds service-related configuration values
type ServiceConfig struct {
	GoAPIBaseURL string
}

// LoadServiceConfig loads service configurations from environment variables
func LoadServiceConfig() ServiceConfig {
	baseURL := os.Getenv("GO_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api"
	}

	return ServiceConfig{
		GoAPIBaseURL: baseURL,
	}
}

// GetGoAPIBaseURL retrieves the GO_API_BASE_URL environment variable with a default fallback
func GetGoAPIBaseURL() string {
	baseURL := os.Getenv("GO_API_BASE_URL")
	if baseURL == "" {
		return "http://localhost:8080/api"
	}
	return baseURL
}

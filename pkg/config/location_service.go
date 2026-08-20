package config

import (
	"strings"
	"time"

	"kota-siaga/utils"
)

const (
	defaultLocationServiceBaseURL = "https://indonesia.imyourz.com"
	defaultLocationServiceTimeout = 5 * time.Second
)

type LocationServiceConfig struct {
	BaseURL string
	Timeout time.Duration
}

func LoadLocationServiceConfig() LocationServiceConfig {
	baseURL := utils.GetEnv("LOCATION_SERVICE_BASE_URL", defaultLocationServiceBaseURL)
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultLocationServiceBaseURL
	}

	return LocationServiceConfig{
		BaseURL: strings.TrimSpace(baseURL),
		Timeout: utils.DurationFromEnv([]string{"LOCATION_SERVICE_TIMEOUT"}, defaultLocationServiceTimeout),
	}
}

func ValidateLocationServiceBaseURL(rawURL string) error {
	return validateHTTPBaseURL(rawURL, "LOCATION_SERVICE_BASE_URL must be a valid URL")
}

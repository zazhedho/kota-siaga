package config

import (
	"strings"
	"time"

	"kota-siaga/utils"
)

const (
	defaultBMKGBaseURL = "https://data.bmkg.go.id"
	defaultBMKGTimeout = 5 * time.Second
)

type BMKGConfig struct {
	BaseURL string
	Timeout time.Duration
}

func LoadBMKGConfig() BMKGConfig {
	baseURL := utils.GetEnv("BMKG_BASE_URL", defaultBMKGBaseURL)
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBMKGBaseURL
	}

	return BMKGConfig{
		BaseURL: strings.TrimSpace(baseURL),
		Timeout: utils.DurationFromEnv([]string{"BMKG_TIMEOUT"}, defaultBMKGTimeout),
	}
}

func ValidateBMKGBaseURL(rawURL string) error {
	return validateHTTPBaseURL(rawURL, "BMKG_BASE_URL must be a valid URL")
}

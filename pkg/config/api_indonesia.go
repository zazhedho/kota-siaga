package config

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"kota-siaga/utils"
)

const (
	defaultAPIIndonesiaBaseURL = "https://use.apiindonesia.id"
	defaultAPIIndonesiaTimeout = 5 * time.Second
)

type APIIndonesiaConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func LoadAPIIndonesiaConfig() APIIndonesiaConfig {
	baseURL := utils.GetEnv("API_INDONESIA_BASE_URL", defaultAPIIndonesiaBaseURL)
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultAPIIndonesiaBaseURL
	}

	return APIIndonesiaConfig{
		BaseURL: strings.TrimSpace(baseURL),
		APIKey:  utils.GetEnv("API_INDONESIA_KEY", ""),
		Timeout: utils.DurationFromEnv([]string{"API_INDONESIA_TIMEOUT"}, defaultAPIIndonesiaTimeout),
	}
}

func ValidateAPIIndonesiaBaseURL(rawURL string) error {
	return validateHTTPBaseURL(rawURL, "API_INDONESIA_BASE_URL must be a valid URL")
}

func validateHTTPBaseURL(rawURL, message string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Hostname() == "" ||
		(!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) ||
		parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return errors.New(message)
	}
	return nil
}

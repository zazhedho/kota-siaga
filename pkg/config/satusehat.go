package config

import (
	"errors"
	"strings"
	"time"

	"kota-siaga/utils"
)

const (
	defaultSATUSEHATBaseURL     = "https://api-satusehat.kemkes.go.id/masterdata"
	defaultSATUSEHATAuthBaseURL = "https://api-satusehat.kemkes.go.id/oauth2/v1"
	defaultSATUSEHATTimeout     = 5 * time.Second
)

type SATUSEHATConfig struct {
	BaseURL      string
	AuthBaseURL  string
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
}

func LoadSATUSEHATConfig() SATUSEHATConfig {
	baseURL := utils.GetEnv("SATUSEHAT_BASE_URL", defaultSATUSEHATBaseURL)
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultSATUSEHATBaseURL
	}
	authBaseURL := utils.GetEnv("SATUSEHAT_AUTH_BASE_URL", defaultSATUSEHATAuthBaseURL)
	if strings.TrimSpace(authBaseURL) == "" {
		authBaseURL = defaultSATUSEHATAuthBaseURL
	}

	return SATUSEHATConfig{
		BaseURL:      strings.TrimSpace(baseURL),
		AuthBaseURL:  strings.TrimSpace(authBaseURL),
		ClientID:     strings.TrimSpace(utils.GetEnv("SATUSEHAT_CLIENT_ID", "")),
		ClientSecret: strings.TrimSpace(utils.GetEnv("SATUSEHAT_CLIENT_SECRET", "")),
		Timeout:      utils.DurationFromEnv([]string{"SATUSEHAT_TIMEOUT"}, defaultSATUSEHATTimeout),
	}
}

func ValidateSATUSEHATBaseURL(rawURL string) error {
	return validateHTTPBaseURL(rawURL, "SATUSEHAT_BASE_URL must be a valid URL")
}

func ValidateSATUSEHATAuthBaseURL(rawURL string) error {
	return validateHTTPBaseURL(rawURL, "SATUSEHAT_AUTH_BASE_URL must be a valid URL")
}

func ValidateSATUSEHATConfig(clientConfig SATUSEHATConfig) error {
	var problems []string
	if strings.TrimSpace(clientConfig.ClientID) == "" {
		problems = append(problems, "SATUSEHAT_CLIENT_ID is required")
	}
	if strings.TrimSpace(clientConfig.ClientSecret) == "" {
		problems = append(problems, "SATUSEHAT_CLIENT_SECRET is required")
	}
	if err := ValidateSATUSEHATBaseURL(clientConfig.BaseURL); err != nil {
		problems = append(problems, err.Error())
	}
	if err := ValidateSATUSEHATAuthBaseURL(clientConfig.AuthBaseURL); err != nil {
		problems = append(problems, err.Error())
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

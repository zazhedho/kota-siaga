package config

import (
	"errors"
	"strconv"
	"strings"

	"kota-siaga/utils"

	"github.com/redis/go-redis/v9"
)

func ValidateStartupConfig(port string) error {
	var problems []string

	problems = append(problems, validateRequiredPort(port)...)
	problems = append(problems, validateAPIIndonesiaConfig()...)
	problems = append(problems, validateBMKGConfig()...)
	problems = append(problems, validateLocationServiceConfig()...)
	problems = append(problems, validateOptionalRedisConfig()...)

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func validateBMKGConfig() []string {
	if err := ValidateBMKGBaseURL(LoadBMKGConfig().BaseURL); err != nil {
		return []string{"BMKG_BASE_URL must be a valid URL"}
	}
	return nil
}

func validateRequiredPort(port string) []string {
	port = strings.TrimSpace(port)
	if port == "" {
		return []string{"PORT is required"}
	}
	if !validPort(port) {
		return []string{"PORT must be a number between 1 and 65535"}
	}
	return nil
}

func validateAPIIndonesiaConfig() []string {
	var problems []string

	if strings.TrimSpace(utils.GetEnv("API_INDONESIA_KEY", "")) == "" {
		problems = append(problems, "API_INDONESIA_KEY is required")
	}

	baseURL := LoadAPIIndonesiaConfig().BaseURL
	if err := ValidateAPIIndonesiaBaseURL(baseURL); err != nil {
		problems = append(problems, "API_INDONESIA_BASE_URL must be a valid URL")
	}

	return problems
}

func validateLocationServiceConfig() []string {
	if err := ValidateLocationServiceBaseURL(LoadLocationServiceConfig().BaseURL); err != nil {
		return []string{"LOCATION_SERVICE_BASE_URL must be a valid URL"}
	}
	return nil
}

func validateOptionalRedisConfig() []string {
	if !hasAnyEnv("REDIS_URL", "REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB") {
		return nil
	}

	var problems []string
	if rawURL := utils.GetEnv("REDIS_URL", ""); rawURL != "" {
		if _, err := redis.ParseURL(rawURL); err != nil {
			problems = append(problems, "REDIS_URL is invalid")
		}
	}
	if port := utils.GetEnv("REDIS_PORT", ""); port != "" && !validPort(port) {
		problems = append(problems, "REDIS_PORT must be a number between 1 and 65535")
	}
	if db := utils.GetEnv("REDIS_DB", ""); db != "" {
		if parsed, err := strconv.Atoi(db); err != nil || parsed < 0 {
			problems = append(problems, "REDIS_DB must be a non-negative integer")
		}
	}
	return problems
}

func validPort(value string) bool {
	port, err := strconv.Atoi(value)
	return err == nil && port >= 1 && port <= 65535
}

func hasAnyEnv(keys ...string) bool {
	for _, key := range keys {
		if utils.GetEnv(key, "") != "" {
			return true
		}
	}
	return false
}

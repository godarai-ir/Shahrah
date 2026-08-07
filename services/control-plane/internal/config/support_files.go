package config

import (
	"os"
	"strings"
	"time"
)

// SupportFileSettings contains immutable private object-storage and scanning settings.
type SupportFileSettings struct {
	Endpoint       string
	Bucket         string
	AccessKey      string
	SecretKey      string
	Region         string
	ForcePathStyle bool
	ClamAVAddress  string
	ClamAVTimeout  time.Duration
	AllowUnscanned bool
}

// LoadSupportFileSettings reads support attachment settings at the configuration edge.
func LoadSupportFileSettings() SupportFileSettings {
	return SupportFileSettings{
		Endpoint:       firstEnvironment("PRIVATE_FILES_S3_ENDPOINT", "SUPPORT_S3_ENDPOINT"),
		Bucket:         firstEnvironment("PRIVATE_FILES_S3_BUCKET", "SUPPORT_S3_BUCKET"),
		AccessKey:      firstEnvironment("PRIVATE_FILES_S3_ACCESS_KEY", "SUPPORT_S3_ACCESS_KEY"),
		SecretKey:      firstEnvironment("PRIVATE_FILES_S3_SECRET_KEY", "SUPPORT_S3_SECRET_KEY"),
		Region:         firstEnvironmentOrDefault("us-east-1", "PRIVATE_FILES_S3_REGION", "SUPPORT_S3_REGION"),
		ForcePathStyle: firstBooleanEnvironment(true, "PRIVATE_FILES_S3_FORCE_PATH_STYLE", "SUPPORT_S3_FORCE_PATH_STYLE"),
		ClamAVAddress:  strings.TrimSpace(os.Getenv("SUPPORT_CLAMAV_ADDRESS")),
		ClamAVTimeout:  durationEnvironment("SUPPORT_CLAMAV_TIMEOUT", 30*time.Second),
		AllowUnscanned: booleanEnv("SUPPORT_ATTACHMENT_ALLOW_UNSCANNED_DEV", false),
	}
}

func firstEnvironment(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	return ""
}

func firstEnvironmentOrDefault(fallback string, keys ...string) string {
	if value := firstEnvironment(keys...); value != "" {
		return value
	}

	return fallback
}

func firstBooleanEnvironment(fallback bool, keys ...string) bool {
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return booleanEnv(key, fallback)
		}
	}

	return fallback
}

func durationEnvironment(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fallback
	}

	return duration
}

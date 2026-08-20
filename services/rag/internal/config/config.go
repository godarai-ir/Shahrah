// Package config loads the managed RAG service runtime configuration.
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/rag/internal/files"
)

// Config is shared by the api and worker deployment modes.
type Config struct {
	Mode                  string
	Address               string
	MetadataURL           string
	VectorURL             string
	RedisURL              string
	GatewayURL            string
	GatewayAPIKey         string
	GatewayRerankModel    string
	GatewayTimeout        time.Duration
	EmbeddingBaseURL      string
	EmbeddingAPIKey       string
	EmbeddingModel        string
	EmbeddingRevision     string
	EmbeddingDimensions   int
	EmbeddingTimeout      time.Duration
	EmbeddingMaxBatchSize int
	TikaURL               string
	InternalToken         string
	SessionSecret         []byte
	ConversationSecret    []byte
	WorkerID              string
	Files                 files.Config
}

// Load reads environment variables and validates fail-closed production defaults.
func Load() (Config, error) {
	gatewayTimeout, err := duration("RAG_GATEWAY_TIMEOUT", 45*time.Second)
	if err != nil {
		return Config{}, err
	}

	embeddingDimensions, err := integer("SHAHRAH_SMART_CACHE_EMBEDDING_DIMENSIONS", 1024)
	if err != nil {
		return Config{}, err
	}

	embeddingTimeout, err := duration("SHAHRAH_SMART_CACHE_EMBEDDING_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	embeddingMaxBatchSize, err := integer("SHAHRAH_SMART_CACHE_EMBEDDING_MAX_BATCH_SIZE", 16)
	if err != nil {
		return Config{}, err
	}

	forcePathStyle, err := boolean("RAG_S3_FORCE_PATH_STYLE", true)
	if err != nil {
		return Config{}, err
	}

	clamAVTimeout, err := duration("RAG_CLAMAV_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Mode:                  env("RAG_MODE", "api"),
		Address:               env("RAG_ADDRESS", ":8090"),
		MetadataURL:           strings.TrimSpace(os.Getenv("DATABASE_URL")),
		VectorURL:             strings.TrimSpace(os.Getenv("RAG_VECTOR_DATABASE_URL")),
		RedisURL:              strings.TrimSpace(os.Getenv("RAG_REDIS_URL")),
		GatewayURL:            env("RAG_GATEWAY_URL", "http://control-plane:4000"),
		GatewayAPIKey:         strings.TrimSpace(os.Getenv("RAG_GATEWAY_API_KEY")),
		GatewayRerankModel:    strings.TrimSpace(os.Getenv("RAG_GATEWAY_RERANK_MODEL")),
		GatewayTimeout:        gatewayTimeout,
		EmbeddingBaseURL:      strings.TrimSpace(os.Getenv("SHAHRAH_SMART_CACHE_EMBEDDING_BASE_URL")),
		EmbeddingAPIKey:       strings.TrimSpace(os.Getenv("SHAHRAH_SMART_CACHE_EMBEDDING_API_KEY")),
		EmbeddingModel:        env("SHAHRAH_SMART_CACHE_EMBEDDING_MODEL", "Bge-m3"),
		EmbeddingRevision:     env("SHAHRAH_SMART_CACHE_EMBEDDING_REVISION", "v1"),
		EmbeddingDimensions:   embeddingDimensions,
		EmbeddingTimeout:      embeddingTimeout,
		EmbeddingMaxBatchSize: embeddingMaxBatchSize,
		TikaURL:               env("RAG_TIKA_URL", "http://tika:9998"),
		InternalToken:         strings.TrimSpace(os.Getenv("RAG_INTERNAL_TOKEN")),
		WorkerID:              env("RAG_WORKER_ID", hostname()),
		Files: files.Config{
			Endpoint:       strings.TrimSpace(os.Getenv("RAG_S3_ENDPOINT")),
			Bucket:         env("RAG_S3_BUCKET", "godarai-rag"),
			AccessKey:      strings.TrimSpace(os.Getenv("RAG_S3_ACCESS_KEY")),
			SecretKey:      strings.TrimSpace(os.Getenv("RAG_S3_SECRET_KEY")),
			Region:         env("RAG_S3_REGION", "us-east-1"),
			ForcePathStyle: forcePathStyle,
			ClamAVAddress:  strings.TrimSpace(os.Getenv("RAG_CLAMAV_ADDRESS")),
			ClamAVTimeout:  clamAVTimeout,
		},
	}

	secret, err := decodeSecret(os.Getenv("RAG_SESSION_SECRET"))
	if err != nil {
		return Config{}, err
	}

	config.SessionSecret = secret

	conversationSecret, err := decodeNamedSecret("RAG_CONVERSATION_SECRET", os.Getenv("RAG_CONVERSATION_SECRET"))
	if err != nil {
		return Config{}, err
	}

	config.ConversationSecret = conversationSecret

	if config.Mode != "api" && config.Mode != "worker" {
		return Config{}, errors.New("RAG_MODE must be api or worker")
	}

	if config.MetadataURL == "" || config.VectorURL == "" || config.RedisURL == "" || config.GatewayAPIKey == "" ||
		config.GatewayRerankModel == "" ||
		config.EmbeddingBaseURL == "" || config.EmbeddingAPIKey == "" {
		return Config{}, errors.New("RAG data stores, gateway rerank model, and smart-cache embedding credentials are required")
	}

	if len(config.InternalToken) < 32 || len(config.SessionSecret) < 32 || len(config.ConversationSecret) < 32 {
		return Config{}, errors.New("RAG_INTERNAL_TOKEN and decoded RAG session secrets must be at least 32 bytes")
	}

	return config, nil
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func duration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}

	return parsed, nil
}

func boolean(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}

	return parsed, nil
}

func integer(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return parsed, nil
}

func decodeSecret(value string) ([]byte, error) {
	return decodeNamedSecret("RAG_SESSION_SECRET", value)
}

func decodeNamedSecret(name, value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("%s is required", name)
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}

	return decoded, nil
}

func hostname() string {
	value, err := os.Hostname()
	if err != nil || strings.TrimSpace(value) == "" {
		return "rag-worker"
	}

	return value
}

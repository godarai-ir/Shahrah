package config

import (
	"os"
	"strings"
	"time"
)

// RAGSettings configures the internal managed RAG data-plane boundary.
type RAGSettings struct {
	URL     string
	Token   string
	Timeout time.Duration
}

// LoadRAGSettings reads optional managed RAG proxy settings.
func LoadRAGSettings() RAGSettings {
	return RAGSettings{
		URL:     strings.TrimRight(strings.TrimSpace(os.Getenv("RAG_SERVICE_URL")), "/"),
		Token:   strings.TrimSpace(os.Getenv("RAG_INTERNAL_TOKEN")),
		Timeout: durationEnvironment("RAG_SERVICE_TIMEOUT", 2*time.Minute),
	}
}

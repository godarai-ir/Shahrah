package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderAccountUsesNoCredential(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		authMode string
		want     bool
	}{
		{
			name:     "internal mock overrides platform managed auth mode",
			provider: "mock",
			authMode: "api_key",
			want:     true,
		},
		{
			name:     "internal mock comparison is normalized",
			provider: " Mock ",
			authMode: "platform_managed",
			want:     true,
		},
		{
			name:     "explicit no credential remains supported",
			provider: "openai",
			authMode: "no_credential",
			want:     true,
		},
		{
			name:     "external provider keeps its configured auth mode",
			provider: "openai",
			authMode: "api_key",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, providerAccountUsesNoCredential(tt.provider, tt.authMode))
		})
	}
}

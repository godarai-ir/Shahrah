package admin

import (
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"github.com/gofiber/fiber/v2"
)

func TestRequirePlatformProviderCredentialsRejectsTypedNilService(t *testing.T) {
	t.Parallel()

	handlers := New(Dependencies{ProviderCredentials: (*platformcredentials.Service)(nil)})

	service, err := handlers.requirePlatformProviderCredentials()
	if service != nil {
		t.Fatalf("service = %v, want nil", service)
	}

	if fiberErr, ok := err.(*fiber.Error); !ok || fiberErr.Code != fiber.StatusServiceUnavailable {
		t.Fatalf("error = %v, want fiber 503", err)
	}
}

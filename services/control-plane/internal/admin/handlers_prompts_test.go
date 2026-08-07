package admin

import (
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/prompts"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestMapPromptStoreErrorReportsMissingMigration(t *testing.T) {
	t.Parallel()

	err := mapPromptStoreError(prompts.ErrStorageUnavailable)

	fiberErr, ok := err.(*fiber.Error)
	require.True(t, ok)
	require.Equal(t, fiber.StatusServiceUnavailable, fiberErr.Code)
	require.Equal(t, "prompt storage is unavailable", fiberErr.Message)
}

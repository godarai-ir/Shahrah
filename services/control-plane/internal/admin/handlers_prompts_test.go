package admin

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestMapPromptStoreErrorReportsMissingMigration(t *testing.T) {
	t.Parallel()

	err := mapPromptStoreError(&pgconn.PgError{Code: "42P01"})

	fiberErr, ok := err.(*fiber.Error)
	require.True(t, ok)
	require.Equal(t, fiber.StatusServiceUnavailable, fiberErr.Code)
	require.Equal(t, "prompt storage migration has not been applied", fiberErr.Message)
}

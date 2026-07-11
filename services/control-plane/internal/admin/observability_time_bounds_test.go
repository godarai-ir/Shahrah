package admin

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyObservabilityTimeBoundsDefaultsToLastDay(t *testing.T) {
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)

	from, to, err := applyObservabilityTimeBounds(nil, nil, now)
	require.NoError(t, err)
	require.NotNil(t, from)
	require.NotNil(t, to)

	assert.Equal(t, now, *to)
	assert.Equal(t, now.Add(-observabilityDefaultRange), *from)
}

func TestApplyObservabilityTimeBoundsRejectsTooWideRange(t *testing.T) {
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	from := now.Add(-(observabilityMaxRange + time.Hour))
	to := now

	_, _, err := applyObservabilityTimeBounds(&from, &to, now)
	require.Error(t, err)

	var fiberErr *fiber.Error
	require.ErrorAs(t, err, &fiberErr)
	assert.Equal(t, fiber.StatusBadRequest, fiberErr.Code)
	assert.Equal(t, "time range must not exceed 2160 hours", fiberErr.Message)
}

func TestApplyObservabilityTimeBoundsAllowsNinetyDayRange(t *testing.T) {
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	from := now.Add(-90 * 24 * time.Hour)
	to := now

	resolvedFrom, resolvedTo, err := applyObservabilityTimeBounds(&from, &to, now)
	require.NoError(t, err)
	require.NotNil(t, resolvedFrom)
	require.NotNil(t, resolvedTo)
	assert.Equal(t, from, *resolvedFrom)
	assert.Equal(t, to, *resolvedTo)
}

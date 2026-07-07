package logretention

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParseLogRetentionDays(t *testing.T) {
	t.Run("positive integers are limited", func(t *testing.T) {
		days, ok := ParseLogRetentionDays(30)
		require.True(t, ok)
		require.NotNil(t, days)
		assert.Equal(t, 30, *days)
	})

	t.Run("json numbers are supported", func(t *testing.T) {
		days, ok := ParseLogRetentionDays(json.Number("90"))
		require.True(t, ok)
		require.NotNil(t, days)
		assert.Equal(t, 90, *days)
	})

	t.Run("nil and unlimited strings are unlimited", func(t *testing.T) {
		for _, value := range []interface{}{nil, "", "null", " unlimited "} {
			days, ok := ParseLogRetentionDays(value)
			assert.True(t, ok, "value=%v", value)
			assert.Nil(t, days, "value=%v", value)
		}
	})

	t.Run("invalid and non-positive values are ignored", func(t *testing.T) {
		for _, value := range []interface{}{0, -1, "0", "-3", "abc", 1.5, map[string]interface{}{"days": 7}} {
			days, ok := ParseLogRetentionDays(value)
			assert.False(t, ok, "value=%v", value)
			assert.Nil(t, days, "value=%v", value)
		}
	})
}

func TestRetentionCleaner_RunOnce(t *testing.T) {
	errBoom := errors.New("delete failed")
	threeDays := 3
	now := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	store := &fakeRetentionStore{
		policies: []LogRetentionPolicy{
			{TenantID: "tenant_unlimited", RetentionDays: nil},
			{TenantID: "tenant_limited", RetentionDays: &threeDays},
			{TenantID: "tenant_fails", RetentionDays: &threeDays},
			{TenantID: "tenant_after_failure", RetentionDays: &threeDays},
		},
		failures: map[string]error{"tenant_fails": errBoom},
	}
	cleaner := NewRetentionCleaner(Config{
		CleanupEnabled: true,
	}, store, zap.NewNop())
	cleaner.now = func() time.Time { return now }

	err := cleaner.runOnce(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, errBoom)

	require.Len(t, store.deletes, 3)
	assert.Equal(t, "tenant_limited", store.deletes[0].tenantID)
	assert.Equal(t, now.AddDate(0, 0, -3), store.deletes[0].cutoff)
	assert.Equal(t, "tenant_fails", store.deletes[1].tenantID)
	assert.Equal(t, "tenant_after_failure", store.deletes[2].tenantID)
}

type fakeRetentionStore struct {
	policies []LogRetentionPolicy
	failures map[string]error
	deletes  []fakeRetentionDelete
}

type fakeRetentionDelete struct {
	tenantID string
	cutoff   time.Time
}

func (s *fakeRetentionStore) ListLogRetentionPolicies(context.Context) ([]LogRetentionPolicy, error) {
	return s.policies, nil
}

func (s *fakeRetentionStore) DeleteLogDataBefore(_ context.Context, tenantID string, cutoff time.Time) error {
	s.deletes = append(s.deletes, fakeRetentionDelete{tenantID: tenantID, cutoff: cutoff})
	return s.failures[tenantID]
}

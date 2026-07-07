package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceRetentionPoliciesSQL_UsesEffectiveEntitlementPrecedence(t *testing.T) {
	normalized := strings.ToLower(strings.Join(strings.Fields(traceRetentionPoliciesSQL), " "))

	assert.Contains(t, normalized, "from tenants")
	assert.Contains(t, normalized, "where ts.status in ('trialing', 'active', 'past_due')")
	assert.Contains(t, normalized, "where fc.feature_key = $1")
	assert.Contains(t, normalized, "and pe.feature_key = $1")
	assert.Contains(t, normalized, "and o.feature_key = $1")
	assert.Contains(t, normalized, "where o.active = true")
	assert.Contains(t, normalized, "and (o.expires_at is null or o.expires_at > now())")
	assert.Contains(t, normalized, "select tenant_id, value_json, precedence from override_values")
	assert.Contains(t, normalized, "select tenant_id, value_json, precedence from plan_values")
	assert.Contains(t, normalized, "select tenant_id, value_json, precedence from base_values")
	assert.Contains(t, normalized, "order by tenant_id, precedence")
}

func TestTraceRetentionPolicySQL_ResolvesOneTenant(t *testing.T) {
	normalized := strings.ToLower(strings.Join(strings.Fields(traceRetentionPolicySQL), " "))

	assert.Contains(t, normalized, "where tenant_id = $2::uuid")
	assert.Contains(t, normalized, "where fc.feature_key = $1")
	assert.Contains(t, normalized, "and pe.feature_key = $1")
	assert.Contains(t, normalized, "and o.feature_key = $1")
	assert.Contains(t, normalized, "order by tenant_id, precedence")
}

func TestTraceRetentionDoesNotUseClickHouseMutations(t *testing.T) {
	normalized := strings.ToLower(strings.Join(strings.Fields(traceRetentionPolicySQL+" "+traceRetentionPoliciesSQL), " "))

	assert.NotContains(t, normalized, "alter table")
	assert.NotContains(t, normalized, " delete where ")
}

func TestClickHouseTraceRetentionDaysValue(t *testing.T) {
	thirty := 30
	tooLarge := maxClickHouseRetentionDays + 100
	zero := 0

	assert.Equal(t, uint16(0), clickHouseTraceRetentionDaysValue(nil))
	assert.Equal(t, uint16(30), clickHouseTraceRetentionDaysValue(&thirty))
	assert.Equal(t, uint16(maxClickHouseRetentionDays), clickHouseTraceRetentionDaysValue(&tooLarge))
	assert.Equal(t, uint16(0), clickHouseTraceRetentionDaysValue(&zero))
}

package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceRetentionPoliciesSQL_UsesEffectiveEntitlementPrecedence(t *testing.T) {
	normalized := strings.ToLower(strings.Join(strings.Fields(traceRetentionPoliciesSQL), " "))

	assert.Contains(t, normalized, "from tenants")
	assert.Contains(t, normalized, "where ts.status in ('active', 'past_due')")
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

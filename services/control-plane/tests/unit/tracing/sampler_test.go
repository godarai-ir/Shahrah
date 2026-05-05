package tracing_test

import (
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/stretchr/testify/assert"
)

func TestSamplerAlwaysSamplesErrorsAndSlowRequests(t *testing.T) {
	s := tracing.NewSampler(0.0, 1500)
	assert.True(t, s.ShouldSample(tracing.DecisionInput{StatusCode: 500, DurationMS: 10}))
	assert.True(t, s.ShouldSample(tracing.DecisionInput{StatusCode: 200, DurationMS: 1800}))
}

func TestSamplerDeterministicOnTenantAndRequest(t *testing.T) {
	s := tracing.NewSampler(0.25, 1500)
	in := tracing.DecisionInput{TenantID: "t1", RequestID: "r1", StatusCode: 200, DurationMS: 100}
	a := s.ShouldSample(in)
	b := s.ShouldSample(in)
	assert.Equal(t, a, b)
}

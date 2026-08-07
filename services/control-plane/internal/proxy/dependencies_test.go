package proxy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBaseHandlerDependenciesReusesConfiguredPipeline(t *testing.T) {
	t.Parallel()

	pipeline := NewRequestPipelineWithConfig(
		nil,
		nil,
		&testBudgeter{},
		BillingCapabilities{},
		nil,
		nil,
		RequestPipelineConfig{DetachedWorkTimeout: 12 * time.Second},
	)
	dependencies := BaseHandlerDependencies{Pipeline: pipeline}

	require.Same(t, pipeline, dependencies.requestPipeline())
	require.Same(t, dependencies.requestPipeline(), dependencies.requestPipeline())
}

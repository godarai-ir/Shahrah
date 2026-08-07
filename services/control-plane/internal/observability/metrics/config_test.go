package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDLQConfigUsesProductionStreamDefaults(t *testing.T) {
	cfg := DefaultDLQConfig()

	assert.Equal(t, "disabled", cfg.Backend)
	assert.Equal(t, defaultDLQStream, cfg.Stream)
	assert.Equal(t, defaultPoisonStream, cfg.PoisonStream)
	assert.Equal(t, int64(10_000), cfg.MaxLen)
	assert.Equal(t, 7*24*time.Hour, cfg.Retention)
	assert.Equal(t, 250*time.Millisecond, cfg.EnqueueTimeout)
}

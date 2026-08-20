package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTypedEnvironmentValuesFailClosed(t *testing.T) {
	t.Run("duration", func(t *testing.T) {
		t.Setenv("RAG_TEST_DURATION", "not-a-duration")

		_, err := duration("RAG_TEST_DURATION", time.Second)
		require.ErrorContains(t, err, "RAG_TEST_DURATION")
	})

	t.Run("boolean", func(t *testing.T) {
		t.Setenv("RAG_TEST_BOOLEAN", "sometimes")

		_, err := boolean("RAG_TEST_BOOLEAN", true)
		require.ErrorContains(t, err, "RAG_TEST_BOOLEAN")
	})

	t.Run("integer", func(t *testing.T) {
		t.Setenv("RAG_TEST_INTEGER", "0")

		_, err := integer("RAG_TEST_INTEGER", 10)
		require.ErrorContains(t, err, "RAG_TEST_INTEGER")
	})
}

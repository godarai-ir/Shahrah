package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupportCursorRoundTrip(t *testing.T) {
	cursor := encodeSupportCursor(50)
	offset, err := decodeSupportCursor(cursor)
	require.NoError(t, err)
	assert.Equal(t, 50, offset)
}

func TestSupportCursorRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"not-base64!", "d3Jvbmc6MQ", encodeSupportCursor(-1)} {
		_, err := decodeSupportCursor(value)
		assert.Error(t, err)
	}
}

func TestParseSupportLimitBounds(t *testing.T) {
	assert.Equal(t, 25, parseSupportLimit(0))
	assert.Equal(t, 10, parseSupportLimit(10))
	assert.Equal(t, 100, parseSupportLimit(500))
}

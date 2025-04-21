package processor

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewLimiter_Success(t *testing.T) {
	ctx := t.Context()

	limiter, promConfig, err := NewLimiter(ctx)

	require.NoError(t, err, "NewLimiter should not return an error during successful setup")

	assert.NotNil(t, limiter, "Returned limiter should not be nil")
	assert.NotNil(t, promConfig, "Returned prometheus config should not be nil")
}

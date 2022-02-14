package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContains(t *testing.T) {
	slice := []string{"A", "B"}
	require.True(t, Contains(slice, "A"))
	require.False(t, Contains(slice, "C"))
}

func TestGetHash(t *testing.T) {
	val, err := GetHash("abcdefgh")
	require.NoError(t, err)
	require.Equal(t, "3ad351775b4634b7", val)
}

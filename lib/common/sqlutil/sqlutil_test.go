package sqlutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlaceholders(t *testing.T) {
	require.Equal(t, "?, ?, ?", Placeholders(3))
	require.Empty(t, Placeholders(0))
	require.Equal(t, "(?, ?), (?, ?)", Rows(2, 2))
}

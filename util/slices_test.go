package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Deletes all consecutive duplicate elements, like Unix uniq.
func TestUniq(t *testing.T) {
	require := require.New(t)
	require.Equal([]int{1, 2, 3}, Uniq([]int{1, 2, 2, 3}))
	require.Equal([]int{1, 2, 3}, Uniq([]int{1, 2, 3, 3}))
	require.Equal([]int{1, 2, 3, 2}, Uniq([]int{1, 1, 2, 3, 3, 2}))
	require.Equal([]int{1}, Uniq([]int{1}))
	require.Equal([]int{1}, Uniq([]int{1, 1}))
	require.Equal([]int{}, Uniq([]int{}))
}

package collection

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFolds(t *testing.T) {
	items := []int{1, 2, 3, 4}
	require.Equal(t, []string{"1", "2", "3", "4"}, Map(items, strconv.Itoa))
	require.Equal(t, []int{2, 4}, Filter(items, func(i int) bool { return i%2 == 0 }))
	require.Equal(t, 10, Reduce(items, 0, func(sum, i int) int { return sum + i }))
	require.Equal(t, map[bool][]int{true: {2, 4}, false: {1, 3}}, GroupBy(items, func(i int) bool { return i%2 == 0 }))
	require.Equal(t, map[int]int{1: 1, 0: 4}, KeyBy([]int{1, 4}, func(i int) int { return i % 2 }))

	_, err := MapErr([]string{"1", "x"}, strconv.Atoi)
	require.Error(t, err)
}

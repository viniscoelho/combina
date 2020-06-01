package types

import (
	"math/rand"
)

// pickRandomValues randomly chooses a number from an slice.
// The number is then removed and returned, along with the
// modified slice.
func pickRandomValue(cur []int) ([]int, int) {
	size := len(cur)
	pos := rand.Intn(size)

	cur[size-1], cur[pos] = cur[pos], cur[size-1]
	return cur[:size-1], cur[size-1]
}

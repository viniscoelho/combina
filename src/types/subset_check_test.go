package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDedupIndex_ExactSameSizeDuplicate(t *testing.T) {
	r := require.New(t)

	existing := [][]int{{3, 1, 2}}
	di := newDedupIndex(existing, 3)

	r.True(di.isDuplicate([]int{1, 2, 3}), "identical same-size game must be rejected")
	r.False(di.isDuplicate([]int{1, 2, 4}), "different same-size game must be allowed")
}

func TestDedupIndex_SubsetOfLongerGame(t *testing.T) {
	r := require.New(t)

	// existing 4-number game; generating 3-number games
	existing := [][]int{{1, 2, 3, 4}}
	di := newDedupIndex(existing, 3)

	r.True(di.isDuplicate([]int{1, 2, 3}), "subset of a longer game must be rejected")
	r.True(di.isDuplicate([]int{2, 3, 4}), "subset of a longer game must be rejected")
	r.False(di.isDuplicate([]int{1, 2, 5}), "non-subset must be allowed")
}

func TestDedupIndex_ShorterExistingNeverBlocks(t *testing.T) {
	r := require.New(t)

	// existing game is smaller than the games being generated: a larger new
	// game can never be a subset of it, so it must never block.
	existing := [][]int{{1, 2}}
	di := newDedupIndex(existing, 3)

	r.False(di.isDuplicate([]int{1, 2, 3}))
	r.False(di.isDuplicate([]int{1, 2, 4}))
}

func TestDedupIndex_RecordBlocksFutureCandidate(t *testing.T) {
	r := require.New(t)

	di := newDedupIndex(nil, 3)
	r.False(di.isDuplicate([]int{1, 2, 3}))

	di.record([]int{1, 2, 3})
	r.True(di.isDuplicate([]int{1, 2, 3}), "recorded game must be rejected afterwards")
}

func TestDedupIndex_MixedExistingSizes(t *testing.T) {
	r := require.New(t)

	existing := [][]int{
		{1, 2, 3},    // same size -> exact
		{4, 5, 6, 7}, // longer -> subset source
		{9, 10},      // shorter -> ignored
	}
	di := newDedupIndex(existing, 3)

	r.True(di.isDuplicate([]int{1, 2, 3}), "exact same-size match")
	r.True(di.isDuplicate([]int{5, 6, 7}), "subset of longer game")
	r.False(di.isDuplicate([]int{4, 5, 8}), "not fully contained anywhere")
}

func TestIsSubset(t *testing.T) {
	r := require.New(t)

	set := map[int]bool{1: true, 2: true, 3: true, 4: true}
	r.True(isSubset([]int{1, 2, 3}, set))
	r.True(isSubset([]int{4}, set))
	r.False(isSubset([]int{1, 5}, set))
}

// TestGenerator_NeverProducesSubsetOfExisting exercises the full generator path
// to ensure a generated game is never a subset of a longer existing game.
func TestGenerator_NeverProducesSubsetOfExisting(t *testing.T) {
	r := require.New(t)

	numGames := 50
	numEach := 15
	gameType := "Lotofacil"

	// One existing 16-number Lotofacil game. Any generated 15-number game that
	// is fully contained in it must be rejected.
	existing := [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}

	dto := newLottoInputDTO(numGames, numEach, []int{}, []int{}, gameType)
	input, err := NewLottoInput(dto)
	r.NoError(err)

	rgg := NewRandomGameGenerator(input, existing)
	lotto := rgg.GenerateLottoCombination()

	super := map[int]bool{}
	for _, n := range existing[0] {
		super[n] = true
	}
	for _, game := range lotto.Numbers.Combination {
		r.False(isSubset(game, super), "generated game must not be a subset of existing longer game: %+v", game)
	}
}

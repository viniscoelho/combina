package types

import (
	"fmt"
	"sort"
)

// dedupIndex rejects a newly generated game when it duplicates or is fully
// contained in a previously seen game. It is shared by every generator.
//
// Two checks run, cheapest first:
//
//   - exact match against games of the SAME size, via a string-hash map (O(1));
//   - subset match against games that are STRICTLY LONGER than the game being
//     generated (a shorter existing game can never contain the new one).
//
// A game longer than the one being produced is only relevant as a potential
// superset, so its numbers are pre-loaded into a set once at construction and
// the per-candidate check is a plain membership test — no map rebuilding per
// generated game.
type dedupIndex struct {
	generated   map[string]bool
	supersets   []map[int]bool
	numEachGame int
}

// newDedupIndex seeds the index from the existing games. Games equal in length
// to numEachGame are hashed for exact-match rejection; games longer than
// numEachGame are stored as membership sets for subset rejection. Shorter games
// are ignored because a new game can never be a subset of a smaller set.
func newDedupIndex(existing [][]int, numEachGame int) *dedupIndex {
	di := &dedupIndex{
		generated:   make(map[string]bool),
		numEachGame: numEachGame,
	}

	for _, game := range existing {
		if len(game) == numEachGame {
			g := make([]int, len(game))
			copy(g, game)
			sort.Ints(g)
			di.generated[fmt.Sprintf("%+v", g)] = true
			continue
		}
		if len(game) > numEachGame {
			set := make(map[int]bool, len(game))
			for _, num := range game {
				set[num] = true
			}
			di.supersets = append(di.supersets, set)
		}
	}

	return di
}

// isDuplicate reports whether sorted is either an exact duplicate of a
// same-size game or a subset of a longer game. sorted must already be sorted
// ascending and have length numEachGame.
func (di *dedupIndex) isDuplicate(sorted []int) bool {
	if di.generated[fmt.Sprintf("%+v", sorted)] {
		return true
	}
	for _, set := range di.supersets {
		if isSubset(sorted, set) {
			return true
		}
	}
	return false
}

// record marks sorted as generated so future candidates are rejected against
// it. sorted must already be sorted ascending.
func (di *dedupIndex) record(sorted []int) {
	di.generated[fmt.Sprintf("%+v", sorted)] = true
}

// isSubset reports whether every number in candidate is present in set.
func isSubset(candidate []int, set map[int]bool) bool {
	for _, num := range candidate {
		if !set[num] {
			return false
		}
	}
	return true
}

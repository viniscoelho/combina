package types

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

type randomGameGenerator struct {
	// Rejects games that duplicate or are subsets of already-seen games
	dedup *dedupIndex
	// A map to count how many times a number has been used (decremented each pick)
	repeated map[int]int
	// Counts how many times each number has been picked across all generated games
	picked map[int]int
	// A slice having the fixed numbers
	fixedNumbers []int
	// A slice having numbers that must never appear in any game
	excludedNumbers []int
	// Number of games (a.k.a. combinations) to be generated
	numGames int
	// Amount of numbers for each game
	numEachGame int
	// Maximum value allowed for each non-fixed number to be used
	maxUsage int
	// Game set kind, e.g., Quina, Lotofacil, ...
	gameType string
	// Minimum and maximum allowed values for a game, e.g, [1, 80]
	gameRange MinMaxRange
	// An alias for the game set
	alias string
}

// NewRandomGameGenerator creates a generator that produces lottery combinations
// with adaptive inverse-frequency selection: numbers picked less often so far
// are preferred, so the full set converges toward an even distribution across
// all numbers.
//
// existing pre-seeds the duplicate-check map so that games already known
// (e.g. imported from a file) are never re-generated.
func NewRandomGameGenerator(input LottoInput, existing [][]int) *randomGameGenerator {
	rgg := randomGameGenerator{}

	rgg.repeated = make(map[int]int)
	rgg.picked = make(map[int]int)
	rgg.fixedNumbers = make([]int, len(input.FixedNumbers))
	rgg.excludedNumbers = make([]int, len(input.ExcludedNumbers))

	copy(rgg.fixedNumbers, input.FixedNumbers)
	copy(rgg.excludedNumbers, input.ExcludedNumbers)

	rgg.numGames = input.NumGames
	rgg.numEachGame = input.NumEachGame
	rgg.gameRange = Games[input.GameType]
	numFixed := len(rgg.fixedNumbers)
	numExcluded := len(rgg.excludedNumbers)
	maxRange := rgg.gameRange.Max

	// maxUsage is the ceiling of (non-fixed slots needed across all games)
	// divided by (non-fixed, non-excluded numbers available). The +1 accounts for integer
	// truncation, and the extra +2 added in initialize() provides a small
	// buffer so the last games are not starved.
	poolSize := maxRange - numFixed - numExcluded
	rgg.maxUsage = ((rgg.numEachGame-numFixed)*rgg.numGames)/poolSize + 1
	if ((rgg.numEachGame-numFixed)*rgg.numGames)%poolSize != 0 {
		rgg.maxUsage++
	}

	rgg.gameType = input.GameType
	rgg.alias = input.Alias

	rgg.initialize()

	rgg.dedup = newDedupIndex(existing, rgg.numEachGame)

	return &rgg
}

// initialize (re-)fills the repeated map with the maximum allowed usage count
// for every non-fixed number in the game range, and resets the picked counters.
// Called once at construction and again whenever the pool is exhausted mid-generation.
func (rgg *randomGameGenerator) initialize() {
	fixed := make(map[int]bool)
	for _, num := range rgg.fixedNumbers {
		fixed[num] = true
	}
	excluded := make(map[int]bool)
	for _, num := range rgg.excludedNumbers {
		excluded[num] = true
	}

	rgg.picked = make(map[int]int)

	minRange, maxRange := rgg.gameRange.Min, rgg.gameRange.Max
	for num := minRange; num <= maxRange; num++ {
		if !fixed[num] && !excluded[num] {
			// +2 over maxUsage gives a small safety margin so numbers are not
			// exhausted before the last games can be assembled.
			rgg.repeated[num] = rgg.maxUsage + 2
		}
	}
}

// GenerateCombination builds a candidate game using adaptive inverse-frequency
// weighting. For each eligible number n (not fixed, not excluded, budget > 0):
//
//	weight = maxPicked - picked[n] + 1
//
// where maxPicked is the highest pick count among eligible numbers this call.
// A number never picked yet gets the highest weight; a number already at the
// maximum count gets weight=1. This steers selection toward under-represented
// numbers so that, across the full batch, no number ends up appearing much more
// or less often than any other. In practice over 150 Lotofácil games the gap
// between the most and least frequent number is ~3 appearances rather than ~11
// with plain uniform selection — all 25 numbers end up close to their expected
// 90 appearances (15 numbers per game × 150 games ÷ 25 numbers = 90).
//
// Returns nil when the pool contains fewer numbers than needed, signalling that
// initialize() must be called before retrying.
func (rgg *randomGameGenerator) GenerateCombination() []int {
	need := rgg.numEachGame - len(rgg.fixedNumbers)

	candidates := make([]int, 0)
	for num, budget := range rgg.repeated {
		if budget > 0 {
			candidates = append(candidates, num)
		}
	}

	if len(candidates) < need {
		return nil
	}

	maxPicked := 0
	for _, n := range candidates {
		if rgg.picked[n] > maxPicked {
			maxPicked = rgg.picked[n]
		}
	}

	weight := func(n int) int {
		return maxPicked - rgg.picked[n] + 1
	}

	return weightedSample(candidates, weight, need)
}

// GenerateValidGame produces one unique game by repeatedly calling
// GenerateCombination until a combination is found that has not been generated
// before. It then appends the fixed numbers, sorts the result, records it in
// the duplicate-check map, and decrements the usage counters for each chosen
// non-fixed number.
//
// If the available pool is exhausted (GenerateCombination returns nil), the
// usage counters are reset via initialize() so generation can continue. This
// avoids an infinite loop when numGames is close to the combinatorial maximum.
func (rgg *randomGameGenerator) GenerateValidGame() []int {
	fixed := make(map[int]bool, len(rgg.fixedNumbers))
	for _, num := range rgg.fixedNumbers {
		fixed[num] = true
	}

	for {
		numbers := rgg.GenerateCombination()
		if numbers == nil {
			// pool exhausted — reset usage counters and retry
			rgg.initialize()
			numbers = rgg.GenerateCombination()
		}

		// add the fixed numbers to the result
		for _, k := range rgg.fixedNumbers {
			numbers = append(numbers, k)
		}
		sort.Slice(numbers, func(i, j int) bool {
			return numbers[i] < numbers[j]
		})

		if rgg.dedup.isDuplicate(numbers) {
			continue
		}
		rgg.dedup.record(numbers)

		// Decrement usage counters and increment picked only for non-fixed numbers.
		for _, num := range numbers {
			if !fixed[num] {
				rgg.repeated[num]--
				rgg.picked[num]++
			}
		}
		return numbers
	}
}

// GenerateLottoCombination generates all numGames combinations and packages
// them into a Lotto value with a new UUID and the current timestamp.
func (rgg *randomGameGenerator) GenerateLottoCombination() Lotto {
	combination := make([][]int, 0)
	for i := 0; i < rgg.numGames; i++ {
		numbers := rgg.GenerateValidGame()
		combination = append(combination, numbers)
	}

	id := uuid.New()
	gc := GameCombo{
		Combination: combination,
		Rows:        rgg.numGames,
		Columns:     rgg.numEachGame,
	}

	return Lotto{
		ID:        id.String(),
		Numbers:   gc,
		GameType:  rgg.gameType,
		CreatedOn: time.Now(),
		Alias:     rgg.alias,
	}
}

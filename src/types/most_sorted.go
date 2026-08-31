package types

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

type fisherYatesModified struct {
	// A map to store each combination that has been generated
	generated map[string]bool
	// A map to count how many times a number has been used
	repeated map[int]int
	// A slice having the fixed numbers
	fixedNumbers []int
	// A slice having the most probably sorted numbers
	mostSortedNumbers []int
	// A slice having all the remaining numbers
	remainingNumbers []int
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

// NewMostSortedShuffle creates a generator that biases selection toward a
// user-supplied set of "most sorted" numbers (numbers statistically drawn more
// often in past results). The bias is achieved by giving those numbers a higher
// weighted probability in each pick — see GenerateCombination for details.
//
// existing pre-seeds the duplicate-check map so that games already known
// (e.g. imported from a file) are never re-generated.
func NewMostSortedShuffle(input LottoInput, existing [][]int) *fisherYatesModified {
	fy := fisherYatesModified{}

	fy.generated = make(map[string]bool)
	fy.repeated = make(map[int]int)
	fy.fixedNumbers = make([]int, len(input.FixedNumbers))
	fy.mostSortedNumbers = make([]int, len(input.MostSortedNumbers))

	copy(fy.fixedNumbers, input.FixedNumbers)
	copy(fy.mostSortedNumbers, input.MostSortedNumbers)

	fy.numGames = input.NumGames
	fy.numEachGame = input.NumEachGame
	fy.gameRange = Games[input.GameType]
	numFixed := len(fy.fixedNumbers)
	maxRange := fy.gameRange.Max

	// maxUsage is the ceiling of (non-fixed slots needed across all games)
	// divided by (non-fixed numbers available). See NewRandomGameGenerator for
	// the same calculation.
	fy.maxUsage = ((fy.numEachGame-numFixed)*fy.numGames)/(maxRange-numFixed) + 1
	if ((fy.numEachGame-numFixed)*fy.numGames)%(maxRange-numFixed) != 0 {
		fy.maxUsage++
	}

	fy.gameType = input.GameType
	fy.alias = input.Alias

	fy.initialize()

	// Mark existing games as already generated so they are skipped during
	// production of new games.
	for _, game := range existing {
		g := make([]int, len(game))
		copy(g, game)
		sort.Ints(g)
		fy.generated[fmt.Sprintf("%+v", g)] = true
	}

	return &fy
}

// initialize (re-)populates repeated and remainingNumbers. Most-sorted numbers
// receive a usage budget of maxUsage * 1.75 so they appear more frequently
// across all generated games. Regular (non-fixed, non-most-sorted) numbers
// receive exactly maxUsage. Fixed numbers are excluded entirely because they
// are appended unconditionally by GenerateValidGame.
//
// Called once at construction and again if the pool is exhausted mid-generation.
func (fy *fisherYatesModified) initialize() {
	fixed, mostSorted := make(map[int]bool), make(map[int]bool)
	for _, num := range fy.fixedNumbers {
		fixed[num] = true
	}
	for _, num := range fy.mostSortedNumbers {
		mostSorted[num] = true
	}

	minRange, maxRange := fy.gameRange.Min, fy.gameRange.Max
	fy.remainingNumbers = make([]int, 0)

	for num := minRange; num <= maxRange; num++ {
		_, isFixed := fixed[num]
		_, isMostSorted := mostSorted[num]

		if isMostSorted {
			fy.repeated[num] = int(float64(fy.maxUsage) * 1.75)
		} else if !isFixed && !isMostSorted {
			fy.repeated[num] = fy.maxUsage
			fy.remainingNumbers = append(fy.remainingNumbers, num)
		}
	}
}

// GenerateCombination implements a modified Fisher-Yates selection that gives
// most-sorted numbers a higher probability of being picked at each position.
//
// At each of the m positions to fill, the algorithm decides which pool to draw
// from using a weighted coin flip:
//
//   - probability of drawing from mostSorted (size k): k*p / (k*p + (n-k)*q)
//   - probability of drawing from remaining  (size n-k): (n-k)*q / (k*p + (n-k)*q)
//
// with p=7, q=3 (roughly 70/30 when pools are equal size). Both k and n
// decrease as numbers are consumed, so the probabilities adjust dynamically.
// pickRandomValue performs a single Fisher-Yates step on the chosen pool:
// it swaps a random element to the end and pops it, giving O(1) sampling
// without replacement.
//
// Returns nil if there are not enough numbers in total to fill m positions, or
// if both pools are empty before m positions are filled.
func (fy *fisherYatesModified) GenerateCombination() []int {
	numbersK, numbersNK := make([]int, len(fy.mostSortedNumbers)), make([]int, len(fy.remainingNumbers))
	copy(numbersK, fy.mostSortedNumbers)
	copy(numbersNK, fy.remainingNumbers)

	// m: non-fixed slots to fill per game
	m := fy.numEachGame - len(fy.fixedNumbers)
	// n: total non-fixed numbers available (decrements each pick)
	n := fy.gameRange.Max - len(fy.fixedNumbers)
	// k: most-sorted numbers not yet picked this game (decrements when one is chosen)
	k := len(fy.mostSortedNumbers)
	// p and q are the relative weights for most-sorted vs remaining pools
	p, q := 7, 3

	if len(numbersK)+len(numbersNK) < m {
		return nil
	}

	result := make([]int, m)
	for i := 0; i < m; i++ {
		total := k*p + (n-k)*q
		if total <= 0 || (k == 0 && len(numbersNK) == 0) {
			return nil
		}
		// Force draw from remaining if mostSorted pool is empty; force draw
		// from mostSorted if remaining pool is empty; otherwise use the weighted coin.
		if k > 0 && len(numbersK) > 0 && (len(numbersNK) == 0 || rand.Intn(total) < k*p) {
			numbersK, result[i] = pickRandomValue(numbersK)
			k--
		} else {
			numbersNK, result[i] = pickRandomValue(numbersNK)
		}
		n--
	}

	return result
}

// isValidGame checks that every number in the candidate combination still has
// remaining usage budget (repeated > 0). A combination that would exceed the
// budget of any number is rejected so that no single number dominates the
// full set of generated games.
func (fy *fisherYatesModified) isValidGame(numbers []int) bool {
	for _, num := range numbers {
		if c := fy.repeated[num]; c <= 0 {
			return false
		}
	}
	return true
}

// GenerateValidGame produces one unique game by repeatedly calling
// GenerateCombination until a result passes isValidGame and has not been
// generated before. It then appends fixed numbers, sorts, records the game in
// the duplicate-check map, and decrements usage counters for non-fixed numbers.
//
// If GenerateCombination returns nil (pool exhausted), initialize() resets the
// usage budgets before retrying, preventing an infinite loop when numGames is
// close to the combinatorial maximum.
func (fy *fisherYatesModified) GenerateValidGame() []int {
	fixed := make(map[int]bool, len(fy.fixedNumbers))
	for _, num := range fy.fixedNumbers {
		fixed[num] = true
	}

	var numbers []int
	for {
		numbers = fy.GenerateCombination()
		if numbers == nil || !fy.isValidGame(numbers) {
			if numbers == nil {
				// pool exhausted — reset usage counters and retry
				fy.initialize()
			}
			continue
		}

		// add the fixed numbers to the result
		for _, num := range fy.fixedNumbers {
			numbers = append(numbers, num)
		}
		sort.Slice(numbers, func(i, j int) bool {
			return numbers[i] < numbers[j]
		})

		hashedNumbers := fmt.Sprintf("%+v", numbers)
		if _, ok := fy.generated[hashedNumbers]; ok {
			continue
		}
		fy.generated[hashedNumbers] = true

		// Decrement usage counters only for non-fixed numbers; fixed numbers
		// are not tracked in repeated and must not be touched.
		for _, num := range numbers {
			if !fixed[num] {
				fy.repeated[num]--
			}
		}
		break
	}

	return numbers
}

// GenerateLottoCombination generates all numGames combinations and packages
// them into a Lotto value with a new UUID and the current timestamp.
func (fy *fisherYatesModified) GenerateLottoCombination() Lotto {
	combination := make([][]int, 0)
	for i := 0; i < fy.numGames; i++ {
		numbers := fy.GenerateValidGame()
		combination = append(combination, numbers)
	}

	id := uuid.New()
	gc := GameCombo{
		Combination: combination,
		Rows:        fy.numGames,
		Columns:     fy.numEachGame,
	}

	return Lotto{
		ID:        id.String(),
		Numbers:   gc,
		GameType:  fy.gameType,
		CreatedOn: time.Now(),
		Alias:     fy.alias,
	}
}

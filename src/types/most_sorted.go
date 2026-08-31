package types

import (
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

type mostSortedGenerator struct {
	// Rejects games that duplicate or are subsets of already-seen games
	dedup *dedupIndex
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

// NewMostSortedGenerator creates a generator that biases selection toward a
// user-supplied set of "most sorted" numbers (numbers statistically drawn more
// often in past results). The bias is achieved by giving those numbers a higher
// weighted probability in each pick — see GenerateCombination for details.
//
// existing pre-seeds the duplicate-check map so that games already known
// (e.g. imported from a file) are never re-generated.
func NewMostSortedGenerator(input LottoInput, existing [][]int) *mostSortedGenerator {
	g := mostSortedGenerator{}

	g.repeated = make(map[int]int)
	g.fixedNumbers = make([]int, len(input.FixedNumbers))
	g.mostSortedNumbers = make([]int, len(input.MostSortedNumbers))

	copy(g.fixedNumbers, input.FixedNumbers)
	copy(g.mostSortedNumbers, input.MostSortedNumbers)

	g.numGames = input.NumGames
	g.numEachGame = input.NumEachGame
	g.gameRange = Games[input.GameType]
	numFixed := len(g.fixedNumbers)
	maxRange := g.gameRange.Max

	// maxUsage is the ceiling of (non-fixed slots needed across all games)
	// divided by (non-fixed numbers available). See NewRandomGameGenerator for
	// the same calculation.
	g.maxUsage = ((g.numEachGame-numFixed)*g.numGames)/(maxRange-numFixed) + 1
	if ((g.numEachGame-numFixed)*g.numGames)%(maxRange-numFixed) != 0 {
		g.maxUsage++
	}

	g.gameType = input.GameType
	g.alias = input.Alias

	g.initialize()

	g.dedup = newDedupIndex(existing, g.numEachGame)

	return &g
}

// initialize (re-)populates repeated and remainingNumbers. Most-sorted numbers
// receive a usage budget of maxUsage * 1.75 so they appear more frequently
// across all generated games. Regular (non-fixed, non-most-sorted) numbers
// receive exactly maxUsage. Fixed numbers are excluded entirely because they
// are appended unconditionally by GenerateValidGame.
//
// Called once at construction and again if the pool is exhausted mid-generation.
func (g *mostSortedGenerator) initialize() {
	fixed, mostSorted := make(map[int]bool), make(map[int]bool)
	for _, num := range g.fixedNumbers {
		fixed[num] = true
	}
	for _, num := range g.mostSortedNumbers {
		mostSorted[num] = true
	}

	minRange, maxRange := g.gameRange.Min, g.gameRange.Max
	g.remainingNumbers = make([]int, 0)

	for num := minRange; num <= maxRange; num++ {
		_, isFixed := fixed[num]
		_, isMostSorted := mostSorted[num]

		if isMostSorted {
			g.repeated[num] = int(float64(g.maxUsage) * 1.75)
		} else if !isFixed && !isMostSorted {
			g.repeated[num] = g.maxUsage
			g.remainingNumbers = append(g.remainingNumbers, num)
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
func (g *mostSortedGenerator) GenerateCombination() []int {
	numbersK, numbersNK := make([]int, len(g.mostSortedNumbers)), make([]int, len(g.remainingNumbers))
	copy(numbersK, g.mostSortedNumbers)
	copy(numbersNK, g.remainingNumbers)

	// m: non-fixed slots to fill per game
	m := g.numEachGame - len(g.fixedNumbers)
	// n: total non-fixed numbers available (decrements each pick)
	n := g.gameRange.Max - len(g.fixedNumbers)
	// k: most-sorted numbers not yet picked this game (decrements when one is chosen)
	k := len(g.mostSortedNumbers)
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
func (g *mostSortedGenerator) isValidGame(numbers []int) bool {
	for _, num := range numbers {
		if c := g.repeated[num]; c <= 0 {
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
func (g *mostSortedGenerator) GenerateValidGame() []int {
	fixed := make(map[int]bool, len(g.fixedNumbers))
	for _, num := range g.fixedNumbers {
		fixed[num] = true
	}

	var numbers []int
	for {
		numbers = g.GenerateCombination()
		if numbers == nil || !g.isValidGame(numbers) {
			if numbers == nil {
				// pool exhausted — reset usage counters and retry
				g.initialize()
			}
			continue
		}

		// add the fixed numbers to the result
		for _, num := range g.fixedNumbers {
			numbers = append(numbers, num)
		}
		sort.Slice(numbers, func(i, j int) bool {
			return numbers[i] < numbers[j]
		})

		if g.dedup.isDuplicate(numbers) {
			continue
		}
		g.dedup.record(numbers)

		// Decrement usage counters only for non-fixed numbers; fixed numbers
		// are not tracked in repeated and must not be touched.
		for _, num := range numbers {
			if !fixed[num] {
				g.repeated[num]--
			}
		}
		break
	}

	return numbers
}

// GenerateLottoCombination generates all numGames combinations and packages
// them into a Lotto value with a new UUID and the current timestamp.
func (g *mostSortedGenerator) GenerateLottoCombination() Lotto {
	combination := make([][]int, 0)
	for i := 0; i < g.numGames; i++ {
		numbers := g.GenerateValidGame()
		combination = append(combination, numbers)
	}

	id := uuid.New()
	gc := GameCombo{
		Combination: combination,
		Rows:        g.numGames,
		Columns:     g.numEachGame,
	}

	return Lotto{
		ID:        id.String(),
		Numbers:   gc,
		GameType:  g.gameType,
		CreatedOn: time.Now(),
		Alias:     g.alias,
	}
}

// pickRandomValue randomly chooses a number from a slice. The number is then
// removed and returned, along with the modified slice.
func pickRandomValue(cur []int) ([]int, int) {
	size := len(cur)
	pos := rand.Intn(size)

	cur[size-1], cur[pos] = cur[pos], cur[size-1]
	return cur[:size-1], cur[size-1]
}

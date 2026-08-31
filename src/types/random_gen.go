package types

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

type randomGameGenerator struct {
	// A map to store each combination that has been generated
	generated map[string]bool
	// A map to count how many times a number has been used
	repeated map[int]int
	// A slice having the fixed numbers
	fixedNumbers []int
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

// NewRandomGameGenerator creates a generator that produces uniformly random
// lottery combinations for the given input. Each non-fixed number is allowed
// to appear at most maxUsage times across all generated games, which keeps the
// distribution roughly even.
//
// existing pre-seeds the duplicate-check map so that games already known
// (e.g. imported from a file) are never re-generated.
func NewRandomGameGenerator(input LottoInput, existing [][]int) *randomGameGenerator {
	rgg := randomGameGenerator{}

	rgg.generated = make(map[string]bool)
	rgg.repeated = make(map[int]int)
	rgg.fixedNumbers = make([]int, len(input.FixedNumbers))

	copy(rgg.fixedNumbers, input.FixedNumbers)

	rgg.numGames = input.NumGames
	rgg.numEachGame = input.NumEachGame
	rgg.gameRange = Games[input.GameType]
	numFixed := len(rgg.fixedNumbers)
	maxRange := rgg.gameRange.Max

	// maxUsage is the ceiling of (non-fixed slots needed across all games)
	// divided by (non-fixed numbers available). The +1 accounts for integer
	// truncation, and the extra +2 added in initialize() provides a small
	// buffer so the last games are not starved.
	rgg.maxUsage = ((rgg.numEachGame-numFixed)*rgg.numGames)/(maxRange-numFixed) + 1
	if ((rgg.numEachGame-numFixed)*rgg.numGames)%(maxRange-numFixed) != 0 {
		rgg.maxUsage++
	}

	rgg.gameType = input.GameType
	rgg.alias = input.Alias

	rgg.initialize()

	// Mark existing games as already generated so they are skipped during
	// production of new games.
	for _, game := range existing {
		g := make([]int, len(game))
		copy(g, game)
		sort.Ints(g)
		rgg.generated[fmt.Sprintf("%+v", g)] = true
	}

	return &rgg
}

// initialize (re-)fills the repeated map with the maximum allowed usage count
// for every non-fixed number in the game range. It is called once at
// construction and again whenever the pool is exhausted mid-generation.
func (rgg *randomGameGenerator) initialize() {
	fixed := make(map[int]bool)
	for _, num := range rgg.fixedNumbers {
		fixed[num] = true
	}

	minRange, maxRange := rgg.gameRange.Min, rgg.gameRange.Max
	for num := minRange; num <= maxRange; num++ {
		if _, ok := fixed[num]; !ok {
			// +2 over maxUsage gives a small safety margin so numbers are not
			// exhausted before the last games can be assembled.
			rgg.repeated[num] = rgg.maxUsage + 2
		}
	}
}

// GenerateCombination builds the pool of non-fixed numbers whose usage budget
// has not been exhausted (repeated > 0), shuffles it with Fisher-Yates via
// rand.Shuffle, and returns the first `need` elements as a candidate
// combination. Returns nil when the pool contains fewer numbers than needed,
// signalling that initialize() must be called before retrying.
func (rgg *randomGameGenerator) GenerateCombination() []int {
	numbers := make([]int, 0)
	for num := range rgg.repeated {
		if c := rgg.repeated[num]; c <= 0 {
			continue
		}
		numbers = append(numbers, num)
	}

	need := rgg.numEachGame - len(rgg.fixedNumbers)
	if len(numbers) < need {
		return nil
	}

	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	return numbers[:need]
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

		hashedNumbers := fmt.Sprintf("%+v", numbers)
		if _, ok := rgg.generated[hashedNumbers]; ok {
			continue
		}
		rgg.generated[hashedNumbers] = true

		// Decrement usage counters only for non-fixed numbers; fixed numbers
		// are not tracked in repeated and must not be touched.
		for _, num := range numbers {
			if !fixed[num] {
				rgg.repeated[num]--
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

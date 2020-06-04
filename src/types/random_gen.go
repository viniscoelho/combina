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

func NewRandomGameGenerator(input LottoInput) *randomGameGenerator {
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

	// calculates how many times each number is allowed to be used
	rgg.maxUsage = ((rgg.numEachGame-numFixed)*rgg.numGames)/(maxRange-numFixed) + 1
	if ((rgg.numEachGame-numFixed)*rgg.numGames)%(maxRange-numFixed) != 0 {
		rgg.maxUsage++
	}

	rgg.gameType = input.GameType
	rgg.alias = input.Alias

	rgg.initialize()
	return &rgg
}

func (rgg *randomGameGenerator) initialize() {
	fixed := make(map[int]bool)
	for _, num := range rgg.fixedNumbers {
		fixed[num] = true
	}

	minRange, maxRange := rgg.gameRange.Min, rgg.gameRange.Max
	for num := minRange; num <= maxRange; num++ {
		if _, ok := fixed[num]; !ok {
			rgg.repeated[num] = rgg.maxUsage
		}
	}
}

// GenerateCombination returns a slice of a shuffled array containing
// valid numbers for a combination.
func (rgg *randomGameGenerator) GenerateCombination() []int {
	numbers := make([]int, 0)
	for num := range rgg.repeated {
		// this will guarantee that all generated combinations are valid,
		// i.e., the combination meets the desired criteria
		if c := rgg.repeated[num]; c <= 0 {
			continue
		}

		numbers = append(numbers, num)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	numFixed := len(rgg.fixedNumbers)
	return numbers[:(rgg.numEachGame - numFixed)]
}

// GenerateValidGame validates if a combination returned by generateCombination
// was not previously generated and returns it
func (rgg *randomGameGenerator) GenerateValidGame() []int {
	for {
		numbers := rgg.GenerateCombination()

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

		fixed := make(map[int]bool)
		for _, num := range rgg.fixedNumbers {
			fixed[num] = true
		}
		// count that these chosen numbers were used
		for i := range numbers {
			rgg.repeated[numbers[i]]--
		}
		return numbers
	}
}

func (rgg *randomGameGenerator) GenerateLottoCombination() Lotto {
	combination := make([][]int, 0)
	for i := 0; i < rgg.numGames; i++ {
		numbers := rgg.GenerateValidGame()
		combination = append(combination, numbers)
		// TODO: remove this comment
		// fmt.Fprintf(os.Stdout, "Combo: %v Numbers: %+v\n", i, numbers)
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

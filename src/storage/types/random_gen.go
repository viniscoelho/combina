package types

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

type randomGameGenerator struct {
	generated    map[string]bool
	repeated     map[int]int
	fixedNumbers []int

	numGames    int
	numEachGame int
	maxValue    int
	maxAllowed  int

	gameType string
	alias    string
}

func NewRandomGameGenerator(dto LottoInputDTO) *randomGameGenerator {
	rgg := randomGameGenerator{}

	rgg.generated = make(map[string]bool)
	rgg.repeated = make(map[int]int)
	rgg.fixedNumbers = make([]int, len(dto.FixedNumbers))

	copy(rgg.fixedNumbers, dto.FixedNumbers)

	numFixed := len(rgg.fixedNumbers)
	rgg.numGames = *dto.NumGames
	rgg.numEachGame = *dto.NumEachGame
	rgg.maxValue = Games[*dto.GameType]

	rgg.maxAllowed = ((rgg.numEachGame-numFixed)*rgg.numGames)/(rgg.maxValue-numFixed) + 1
	if ((rgg.numEachGame-numFixed)*rgg.numGames)%(rgg.maxValue-numFixed) != 0 {
		rgg.maxAllowed++
	}

	rgg.gameType = *dto.GameType
	rgg.alias = *dto.Alias

	rgg.initialize()
	return &rgg
}

func (rgg *randomGameGenerator) initialize() {
	fixed := make(map[int]bool)
	for _, num := range rgg.fixedNumbers {
		fixed[num] = true
	}

	lo, hi := 1, rgg.maxValue
	// workaround for Lotomania
	if rgg.maxValue == 100 {
		lo--
		hi--
	}
	for num := lo; num <= hi; num++ {
		if _, ok := fixed[num]; !ok {
			rgg.repeated[num] = rgg.maxAllowed
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

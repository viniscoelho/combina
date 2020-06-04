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

func NewMostSortedShuffle(input LottoInput) *fisherYatesModified {
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

	// calculates how many times each number is allowed to be used
	fy.maxUsage = ((fy.numEachGame-numFixed)*fy.numGames)/(maxRange-numFixed) + 1
	if ((fy.numEachGame-numFixed)*fy.numGames)%(maxRange-numFixed) != 0 {
		fy.maxUsage++
	}

	fy.gameType = input.GameType
	fy.alias = input.Alias

	fy.initialize()
	return &fy
}

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
			fy.repeated[num] = int(float64(fy.maxUsage) * 1.5)
		} else if !isFixed && !isMostSorted {
			fy.repeated[num] = fy.maxUsage
			fy.remainingNumbers = append(fy.remainingNumbers, num)
		}
	}
}

// GenerateCombination picks random values from most sorted
// and remaining slices, thus, generating a combination.
// The numbers from the most sorted slice have a higher probability
// to be included on each combination.
func (fy *fisherYatesModified) GenerateCombination() []int {
	numbersK, numbersNK := make([]int, len(fy.mostSortedNumbers)), make([]int, len(fy.remainingNumbers))
	copy(numbersK, fy.mostSortedNumbers)
	copy(numbersNK, fy.remainingNumbers)

	// numbers within a combination
	m := fy.numEachGame - len(fy.fixedNumbers)
	// numbers allowed to be chosen
	n := fy.gameRange.Max - len(fy.fixedNumbers)
	// numbers that have higher probability to be chosen
	k := len(fy.mostSortedNumbers)
	// probability of a number to be chosen from sets k and nk, respectively
	// the higher the value, the higher the probability
	p, q := 7, 3

	rand.Seed(time.Now().UnixNano())
	result := make([]int, m)
	for i := 0; i < m; i++ {
		if rand.Intn(k*p+(n-k)*q) < k*p {
			numbersK, result[i] = pickRandomValue(numbersK)
			k--
		} else {
			numbersNK, result[i] = pickRandomValue(numbersNK)
		}
		n--
	}

	return result
}

func (fy *fisherYatesModified) isValidGame(numbers []int) bool {
	for _, num := range numbers {
		if c := fy.repeated[num]; c <= 0 {
			return false
		}
	}
	return true
}

func (fy *fisherYatesModified) GenerateValidGame() []int {
	var numbers []int
	for {
		numbers = fy.GenerateCombination()
		if !fy.isValidGame(numbers) {
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

		// mark these chosen numbers as used
		for i := range numbers {
			fy.repeated[numbers[i]]--
		}
		break
	}

	return numbers
}

func (fy *fisherYatesModified) GenerateLottoCombination() Lotto {
	combination := make([][]int, 0)
	for i := 0; i < fy.numGames; i++ {
		numbers := fy.GenerateValidGame()
		combination = append(combination, numbers)
		// TODO: remove this comment
		// fmt.Fprintf(os.Stdout, "Numbers: %v\n", numbers)
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

package types

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

var (
	Games = map[string]MinMaxRange{
		"Lotofacil":    {1, 25},
		"Lotomania":    {0, 99},
		"Quina":        {1, 80},
		"Mega-Sena":    {1, 60},
		"Quina-Brasil": {1, 80},
		"Seninha":      {1, 60},
	}
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

func NewMostSortedShuffle(dto LottoInputDTO) *fisherYatesModified {
	fy := fisherYatesModified{}

	fy.generated = make(map[string]bool)
	fy.repeated = make(map[int]int)
	fy.fixedNumbers = make([]int, len(dto.FixedNumbers))
	fy.mostSortedNumbers = make([]int, len(dto.MostSortedNumbers))

	copy(fy.fixedNumbers, dto.FixedNumbers)
	copy(fy.mostSortedNumbers, dto.MostSortedNumbers)

	fy.numGames = *dto.NumGames
	fy.numEachGame = *dto.NumEachGame
	fy.gameRange = Games[*dto.GameType]
	numFixed := len(fy.fixedNumbers)
	maxRange := fy.gameRange.Max

	// calculates how many times each number is allowed to be used
	fy.maxUsage = ((fy.numEachGame-numFixed)*fy.numGames)/(maxRange-numFixed) + 1
	if ((fy.numEachGame-numFixed)*fy.numGames)%(maxRange-numFixed) != 0 {
		fy.maxUsage++
	}

	fy.gameType = *dto.GameType
	fy.alias = *dto.Alias

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
	numRemaining := maxRange - len(fy.fixedNumbers) - len(fy.mostSortedNumbers)
	fy.remainingNumbers = make([]int, numRemaining)

	for num, cur := minRange, 0; num <= maxRange; num++ {
		_, isFixed := fixed[num]
		_, isMostSorted := mostSorted[num]

		if isMostSorted {
			fy.repeated[num] = int(float64(fy.maxUsage) * 1.5)
		} else if !isFixed && !isMostSorted {
			fy.repeated[num] = fy.maxUsage
			fy.remainingNumbers[cur] = num
			cur++
		}
	}
}

// GenerateCombination picks random values from most sorted
// and remaining slices, thus, generating a combination.
// The numbers from the most sorted slice have a higher probability
// to be included on each combination.
func (fy *fisherYatesModified) GenerateCombination() []int {
	numbers_k, numbers_nk := make([]int, len(fy.mostSortedNumbers)), make([]int, len(fy.remainingNumbers))
	copy(numbers_k, fy.mostSortedNumbers)
	copy(numbers_nk, fy.remainingNumbers)

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
			numbers_k, result[i] = pickRandomValue(numbers_k)
			k--
		} else {
			numbers_nk, result[i] = pickRandomValue(numbers_nk)
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

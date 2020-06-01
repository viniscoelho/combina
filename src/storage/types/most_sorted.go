package types

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

type fisherYatesModified struct {
	generated         map[string]bool
	repeated          map[int]int
	fixedNumbers      []int
	mostSortedNumbers []int
	remainingNumbers  []int

	numGames    int
	numEachGame int
	maxValue    int
	maxRepeated int

	gameType string
	alias    string
}

func NewMostSortedShuffle(dto LottoInputDTO) *fisherYatesModified {
	fy := fisherYatesModified{}

	fy.generated = make(map[string]bool)
	fy.repeated = make(map[int]int)
	fy.fixedNumbers = make([]int, len(dto.FixedNumbers))
	fy.mostSortedNumbers = make([]int, len(dto.MostSortedNumbers))
	fy.remainingNumbers = make([]int, 0)

	copy(fy.fixedNumbers, dto.FixedNumbers)
	copy(fy.mostSortedNumbers, dto.MostSortedNumbers)

	numFixed := len(fy.fixedNumbers)
	fy.numGames = *dto.NumGames
	fy.numEachGame = *dto.NumEachGame
	fy.maxValue = Games[*dto.GameType]

	fy.maxRepeated = ((fy.numEachGame-numFixed)*fy.numGames)/(fy.maxValue-numFixed) + 1
	if ((fy.numEachGame-numFixed)*fy.numGames)%(fy.maxValue-numFixed) != 0 {
		fy.maxRepeated++
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

	lo, hi := 1, fy.maxValue
	// workaround for Lotomania
	if fy.maxValue == 100 {
		lo--
		hi--
	}
	for num := lo; num <= hi; num++ {
		_, isFixed := fixed[num]
		_, isMostSorted := mostSorted[num]
		if !isFixed && !isMostSorted {
			fy.remainingNumbers = append(fy.remainingNumbers, num)
		}

		if !isFixed {
			fy.repeated[num] = fy.maxRepeated
		}
	}
}

// GenerateCombination picks random values from most sorted
// and remaining slices, thus, generating a combination.
// The numbers from the most sorted slice have a higher probability
// to be included on each combination.
func (fy *fisherYatesModified) GenerateCombination() []int {
	rand.Seed(time.Now().UnixNano())

	numbers_k, numbers_nk := make([]int, len(fy.mostSortedNumbers)), make([]int, len(fy.remainingNumbers))
	copy(numbers_k, fy.mostSortedNumbers)
	copy(numbers_nk, fy.remainingNumbers)

	m := fy.numEachGame - len(fy.fixedNumbers)
	n := fy.maxValue - len(fy.fixedNumbers)
	k := len(fy.mostSortedNumbers)
	p, q := 5, 3

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

package routes

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/combina/src/storage/types"
	"github.com/google/uuid"
)

const idRouteVar = "id"

var (
	exists = struct{}{}
)

// isNumEachWithinRange validates if the amount of picked numbers
// is valid according to the official lottery rules
func isNumEachWithinRange(numEachGame int, gameType string) bool {
	switch gameType {
	case "Lotofacil":
		if numEachGame >= 15 && numEachGame <= 18 {
			return true
		}
	case "Lotomania":
		if numEachGame == 50 {
			return true
		}
	case "Quina":
		if numEachGame >= 5 && numEachGame <= 15 {
			return true
		}
	case "Mega-Sena":
		if numEachGame >= 6 && numEachGame <= 15 {
			return true
		}
	case "Quina-Brasil":
		if numEachGame == 13 {
			return true
		}
	case "Seninha":
		if numEachGame == 20 {
			return true
		}
	}
	return false
}

func isValidFixedNumbers(maxValue int, fixedNumbers []int) bool {
	lo, hi := 1, maxValue
	// workaround for Lotomania
	if maxValue == 100 {
		lo--
		hi--
	}

	for _, num := range fixedNumbers {
		if num < lo || num > hi {
			return false
		}
	}
	return true
}

// isValidNumGames validates if the number of games chosen is
// possible to be generated. It follows the combination formula:
// nCr = n!/r!(n-r)!
// n = maxValue-numFixed, r = numPicked-numFixed, c = n-r
func isValidNumGames(numGames int64, maxValue, numPicked, numFixed int) bool {
	n, r := maxValue-numFixed, numPicked-numFixed
	// if it reached this point of validation and r > 20,
	// this means that it is a Lotomania game. Therefore,
	// there is no need to do this calculation, because the
	// number of games will be valid anyway.
	if r > 20 {
		return true
	}

	fact := make([]int64, 21)
	fact[0], fact[1] = 1, 1
	for i := 2; i <= 20; i++ {
		fact[i] = fact[i-1] * int64(i)
	}

	rem := int64(1)
	for i := n - r + 1; i <= n; i++ {
		rem *= int64(i)
	}

	log.Printf("Possible combinations: %v", rem/fact[r])
	if numGames > rem/fact[r] {
		return false
	}

	return true
}

func validateLottoDTO(dto types.LottoInputDTO) error {
	if dto.NumGames == nil || dto.NumEachGame == nil || dto.GameType == nil {
		return types.MissingFieldsError{}
	}

	if _, ok := types.Games[*dto.GameType]; !ok {
		return types.InvalidDTOError{Message: "invalid game type"}
	}

	if *dto.NumGames <= 0 {
		return types.InvalidDTOError{Message: "number of games should be greater than zero"}
	}

	if len(dto.FixedNumbers) > *dto.NumEachGame {
		return types.InvalidDTOError{Message: "amount of fixed numbers cannot be greater than picked numbers"}
	}

	if !isNumEachWithinRange(*dto.NumEachGame, *dto.GameType) {
		return types.InvalidDTOError{Message: "amount of picked numbers should be within a valid range"}
	}

	if !isValidFixedNumbers(types.Games[*dto.GameType], dto.FixedNumbers) {
		return types.InvalidDTOError{Message: "some fixed numbers are invalid -- choose numbers within a valid range"}
	}

	if !isValidNumGames(int64(*dto.NumGames), types.Games[*dto.GameType], *dto.NumEachGame, len(dto.FixedNumbers)) {
		return types.InvalidDTOError{Message: "number of games is invalid -- use another value or change the amount of fixed numbers"}
	}

	if dto.Alias == nil {
		dto.Alias = new(string)
		*dto.Alias = "default"
	}

	return nil
}

func newLottoCombination(dto types.LottoInputDTO) types.Lotto {
	combination := make([][]int, 0)
	fixed := make(map[int]struct{})
	generated := make(map[string]struct{})
	repeated := make(map[int]int)

	for _, num := range dto.FixedNumbers {
		fixed[num] = exists
	}

	numGames := *dto.NumGames
	numFixed := len(fixed)
	numPicked := *dto.NumEachGame
	maxValue := types.Games[*dto.GameType]
	maxRepeated := ((numPicked-numFixed)*numGames)/(maxValue-numFixed) + 1

	if ((numPicked-numFixed)*numGames)%(maxValue-numFixed) != 0 {
		log.Printf("Mod: %v", ((numPicked-numFixed)*numGames)%(maxValue-numFixed))
		maxRepeated++
	}
	log.Printf("Max repetition: %v", maxRepeated)

	for i := 0; i < numGames; i++ {
		numbers := generateValidGame(numPicked, maxValue, maxRepeated, fixed, generated, repeated)
		combination = append(combination, numbers)
		fmt.Fprintf(os.Stdout, "Numbers: %v\n", numbers)
	}

	id := uuid.New()
	gc := types.GameCombo{
		Combination: combination,
		Rows:        numGames,
		Columns:     numPicked,
	}

	return types.Lotto{
		ID:        id.String(),
		Numbers:   gc,
		GameType:  *dto.GameType,
		CreatedOn: time.Now(),
		Alias:     *dto.Alias,
	}
}

// getShuffledNumbers returns a slice of a shuffled array containing
// valid numbers for a combination
func generateCombination(numPicked, maxValue, maxRepeated int, fixedNumbers map[int]struct{}, generated map[string]struct{}, repeated map[int]int) []int {
	lo, hi := 1, maxValue
	// workaround for Lotomania
	if maxValue == 100 {
		lo--
		hi--
	}

	numbers := make([]int, 0)
	for num := lo; num <= hi; num++ {
		if _, ok := fixedNumbers[num]; ok {
			continue
		}

		// this will guarantee that all generated combinations are valid,
		// i.e., the combination meets the desired criteria
		if c := repeated[num]; c == maxRepeated {
			continue
		}

		numbers = append(numbers, num)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(numbers), func(i, j int) { numbers[i], numbers[j] = numbers[j], numbers[i] })

	return numbers[:(numPicked - len(fixedNumbers))]
}

// generateValidGame validates if a combination returned by generateCombination
// was not previously generated and returns it
func generateValidGame(numPicked, maxValue, maxRepeated int, fixedNumbers map[int]struct{}, generated map[string]struct{}, repeated map[int]int) []int {
	var numbers []int
	for {
		numbers = generateCombination(numPicked, maxValue, maxRepeated, fixedNumbers, generated, repeated)

		// add the fixed numbers to the result
		for k := range fixedNumbers {
			numbers = append(numbers, k)
		}
		sort.Slice(numbers, func(i, j int) bool {
			return numbers[i] < numbers[j]
		})

		hashedNumbers := fmt.Sprintf("%+v", numbers)
		if _, ok := generated[hashedNumbers]; ok {
			continue
		}
		generated[hashedNumbers] = exists

		// count that these chosen numbers were used
		for i := range numbers {
			repeated[numbers[i]]++
		}
		break
	}

	return numbers
}

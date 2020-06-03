package routes

import (
	"github.com/combina/src/storage/types"
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

func isValidNumbers(r types.MinMaxRange, numbers []int) bool {
	minRange, maxRange := r.Min, r.Max
	for _, num := range numbers {
		if num < minRange || num > maxRange {
			return false
		}
	}
	return true
}

// isValidNumGames validates if the number of games chosen is
// possible to be generated. It follows the combination formula:
// nCr = n!/r!(n-r)!
// n = maxValue-numFixed, r = numPicked-numFixed, c = n-r
func isValidNumGames(numGames int64, maxRange, numPicked, numFixed int) bool {
	n, r := maxRange-numFixed, numPicked-numFixed
	// if it reached this point of validation and r > 20,
	// this means that it is a Lotomania game. Therefore,
	// there is no need to do this calculation, because the
	// number of games will be always valid.
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

	if numGames > rem/fact[r] {
		return false
	}

	return true
}

func validateInputDTO(dto types.LottoInputDTO) error {
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

	r := types.Games[*dto.GameType]
	if len(dto.MostSortedNumbers) > r.Max-len(dto.FixedNumbers) {
		return types.InvalidDTOError{Message: "amount of most sorted numbers cannot be greater than remaining numbers"}
	}

	if !isNumEachWithinRange(*dto.NumEachGame, *dto.GameType) {
		return types.InvalidDTOError{Message: "amount of picked numbers should be within a valid range"}
	}

	if !isValidNumbers(r, dto.FixedNumbers) {
		return types.InvalidDTOError{Message: "some fixed numbers are invalid -- choose numbers within a valid range"}
	}

	if !isValidNumbers(r, dto.MostSortedNumbers) {
		return types.InvalidDTOError{Message: "some most sorted numbers are invalid -- choose numbers within a valid range"}
	}

	if !isValidNumGames(int64(*dto.NumGames), r.Max, *dto.NumEachGame, len(dto.FixedNumbers)) {
		return types.InvalidDTOError{Message: "number of games is invalid -- use another value or change the amount of fixed numbers"}
	}

	if dto.Alias == nil {
		dto.Alias = new(string)
		*dto.Alias = "default"
	}

	return nil
}

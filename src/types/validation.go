package types

import "fmt"

// IsNumEachValid validates if the amount of picked numbers
// is valid according to the official lottery rules
func IsNumEachValid(numEachGame int, gameType string) bool {
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
	case "Quininha":
		if (numEachGame >= 15 && numEachGame <= 20) || numEachGame == 25 || numEachGame == 30 {
			return true
		}
	case "Seninha":
		if (numEachGame >= 14 && numEachGame <= 20) || numEachGame == 25 || numEachGame == 30 {
			return true
		}
	}
	return false
}

func isValidNumbers(r MinMaxRange, numbers []int) bool {
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
// n = maxValue-numFixed, r = numEachGame-numFixed, c = n-r
func isValidNumGames(numGames int64, maxRange, numEachGame, numFixed int) bool {
	n, r := maxRange-numFixed, numEachGame-numFixed
	// Skip combinatorial check for Lotomania (r > 20) — the number of valid
	// games is always astronomically large. For r > 20 the nCr values exceed
	// int64 range even with the interleaved-divide method.
	if r > 20 {
		return true
	}

	// Compute C(n, r) using the multiplicative formula with interleaved division
	// to avoid int64 overflow: C(n,r) = product_{i=1}^{r} (n-r+i)/i
	// Each intermediate value is an integer because C(n,i) is always integer.
	nCr := int64(1)
	for i := 1; i <= r; i++ {
		nCr = nCr * int64(n-r+i) / int64(i)
		if nCr >= numGames {
			// partial product already at least as large as numGames;
			// the final C(n,r) can only be larger, so the request is valid.
			return true
		}
	}

	return numGames <= nCr
}

// validateWeightedIntersection ensures no number appears in more than one of
// the four classification lists: fixed, excluded, favored, disfavored.
func validateWeightedIntersection(fixed, excluded, favored, disfavored []int) error {
	seen := make(map[int]string, len(fixed)+len(excluded)+len(favored)+len(disfavored))
	for _, n := range fixed {
		seen[n] = "fixed"
	}
	for _, n := range excluded {
		if l, ok := seen[n]; ok {
			return InvalidDTOError{Message: fmt.Sprintf("number %d appears in both excluded and %s", n, l)}
		}
		seen[n] = "excluded"
	}
	for _, n := range favored {
		if l, ok := seen[n]; ok {
			return InvalidDTOError{Message: fmt.Sprintf("number %d appears in both favored and %s", n, l)}
		}
		seen[n] = "favored"
	}
	for _, n := range disfavored {
		if l, ok := seen[n]; ok {
			return InvalidDTOError{Message: fmt.Sprintf("number %d appears in both disfavored and %s", n, l)}
		}
		seen[n] = "disfavored"
	}
	return nil
}

func validateInputDTO(dto LottoInputDTO) error {
	if dto.NumGames == nil || dto.NumEachGame == nil || dto.GameType == nil {
		return MissingFieldsError{}
	}

	if _, ok := Games[*dto.GameType]; !ok {
		return InvalidDTOError{Message: "invalid game type"}
	}

	if *dto.NumGames <= 0 {
		return InvalidDTOError{Message: "number of games should be greater than zero"}
	}

	if len(dto.FixedNumbers) > *dto.NumEachGame {
		return InvalidDTOError{Message: "amount of fixed numbers cannot be greater than picked numbers"}
	}

	r := Games[*dto.GameType]
	if len(dto.FavoredNumbers) > r.Max-len(dto.FixedNumbers) {
		return InvalidDTOError{Message: "amount of favored numbers cannot be greater than remaining numbers"}
	}

	if !IsNumEachValid(*dto.NumEachGame, *dto.GameType) {
		return InvalidDTOError{Message: "amount of picked numbers should be within a valid range"}
	}

	if !isValidNumbers(r, dto.FixedNumbers) {
		return InvalidDTOError{Message: "some fixed numbers are invalid -- choose numbers within a valid range"}
	}

	if !isValidNumbers(r, dto.FavoredNumbers) {
		return InvalidDTOError{Message: "some favored numbers are invalid -- choose numbers within a valid range"}
	}

	if !isValidNumbers(r, dto.DisfavoredNumbers) {
		return InvalidDTOError{Message: "some disfavored numbers are invalid -- choose numbers within a valid range"}
	}

	if !isValidNumbers(r, dto.ExcludedNumbers) {
		return InvalidDTOError{Message: "some excluded numbers are invalid -- choose numbers within a valid range"}
	}

	available := r.Max - r.Min + 1 - len(dto.FixedNumbers) - len(dto.ExcludedNumbers)
	need := *dto.NumEachGame - len(dto.FixedNumbers)
	if available < need {
		return InvalidDTOError{Message: "too many excluded numbers -- not enough numbers left to fill a game"}
	}

	if !isValidNumGames(int64(*dto.NumGames), r.Max, *dto.NumEachGame, len(dto.FixedNumbers)) {
		return InvalidDTOError{Message: "number of games is invalid -- use another value or change the amount of fixed numbers"}
	}

	if err := validateWeightedIntersection(dto.FixedNumbers, dto.ExcludedNumbers, dto.FavoredNumbers, dto.DisfavoredNumbers); err != nil {
		return err
	}

	return nil
}

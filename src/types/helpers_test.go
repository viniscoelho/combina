package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const maxRepetitions = 100000

func newLottoInputDTO(numGames, numEachGame int, fixedNumbers, favoredNumbers []int, gameType string) LottoInputDTO {
	dto := LottoInputDTO{
		NumGames:          &numGames,
		NumEachGame:       &numEachGame,
		FixedNumbers:      make([]int, len(fixedNumbers)),
		FavoredNumbers: make([]int, len(favoredNumbers)),
		GameType:          &gameType,
		Alias:             new(string),
	}

	copy(dto.FixedNumbers, fixedNumbers)
	copy(dto.FavoredNumbers, favoredNumbers)
	*dto.Alias = "test"

	return dto
}

func TestInvalidDTO_Intersection(t *testing.T) {
	r := require.New(t)

	numGames := 100
	numEach := 13
	fixed := []int{13, 41, 60, 78}
	favored := []int{5, 7, 13, 21, 25, 32, 37, 39, 45, 51, 55, 56, 61, 64, 74, 80}
	gameType := "Quina-Brasil"
	dto := newLottoInputDTO(numGames, numEach, fixed, favored, gameType)
	_, err := NewLottoInput(dto)
	r.Error(err)
}

func TestWeightedSample_CorrectCountAndDistinct(t *testing.T) {
	r := require.New(t)

	candidates := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	weight := func(n int) int { return 1 }

	for i := 0; i < 200; i++ {
		result := weightedSample(candidates, weight, 5)
		r.Equal(5, len(result))
		seen := map[int]bool{}
		for _, n := range result {
			r.False(seen[n], "duplicate %d in %v", n, result)
			seen[n] = true
		}
	}
}

func TestWeightedSample_HigherWeightPickedMoreOften(t *testing.T) {
	r := require.New(t)

	// number 1 has weight 100, all others weight 1 — 1 should dominate
	candidates := []int{1, 2, 3, 4, 5}
	weight := func(n int) int {
		if n == 1 {
			return 100
		}
		return 1
	}

	count1 := 0
	trials := 1000
	for i := 0; i < trials; i++ {
		result := weightedSample(candidates, weight, 1)
		if result[0] == 1 {
			count1++
		}
	}
	// number 1 has weight 100/(100+4) ≈ 96% chance; expect > 80%
	r.Greater(count1, trials*80/100, "number 1 picked %d/%d times, expected >80%%", count1, trials)
}

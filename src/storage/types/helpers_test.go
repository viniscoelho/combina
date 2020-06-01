package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const maxRepetitions = 100000

func newValidLottoInputDTO() LottoInputDTO {
	dto := LottoInputDTO{
		NumGames:          new(int),
		NumEachGame:       new(int),
		FixedNumbers:      []int{11, 27, 33},
		MostSortedNumbers: []int{1, 3, 32, 54, 70},
		GameType:          new(string),
		Alias:             new(string),
	}
	*dto.NumGames = 100
	*dto.NumEachGame = 13
	*dto.GameType = "Quina-Brasil"
	*dto.Alias = "test"

	return dto
}

func TestPickRandomValue(t *testing.T) {
	r := require.New(t)
	dto := newValidLottoInputDTO()
	rgg := NewMostSortedShuffle(dto)

	numbers_k, numbers_nk := make([]int, len(rgg.mostSortedNumbers)), make([]int, len(rgg.remainingNumbers))
	copy(numbers_k, rgg.mostSortedNumbers)
	copy(numbers_nk, rgg.remainingNumbers)

	numbers_k, _ = pickRandomValue(numbers_k)
	r.Equal(len(rgg.mostSortedNumbers)-1, len(numbers_k))

	numbers_nk, _ = pickRandomValue(numbers_nk)
	r.Equal(len(rgg.remainingNumbers)-1, len(numbers_nk))
}

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const maxRepetitions = 100000

func newLottoInputDTO(numGames, numEachGame int, fixedNumbers, mostSortedNumbers []int, gameType string) LottoInputDTO {
	dto := LottoInputDTO{
		NumGames:          &numGames,
		NumEachGame:       &numEachGame,
		FixedNumbers:      make([]int, len(fixedNumbers)),
		MostSortedNumbers: make([]int, len(mostSortedNumbers)),
		GameType:          &gameType,
		Alias:             new(string),
	}

	copy(dto.FixedNumbers, fixedNumbers)
	copy(dto.MostSortedNumbers, mostSortedNumbers)
	*dto.Alias = "test"

	return dto
}

func TestInvalidDTO_Intersection(t *testing.T) {
	r := require.New(t)

	numGames := 100
	numEach := 13
	fixed := []int{13, 41, 60, 78}
	mostSorted := []int{5, 7, 13, 21, 25, 32, 37, 39, 45, 51, 55, 56, 61, 64, 74, 80}
	gameType := "Quina-Brasil"
	dto := newLottoInputDTO(numGames, numEach, fixed, mostSorted, gameType)
	_, err := NewLottoInput(dto)
	r.Error(err)
}

func TestPickRandomValue(t *testing.T) {
	r := require.New(t)

	numGames := 100
	numEach := 13
	fixed := []int{13, 41, 60, 78}
	mostSorted := []int{5, 7, 12, 21, 25, 32, 37, 39, 45, 51, 55, 56, 61, 64, 74, 80}
	gameType := "Quina-Brasil"
	dto := newLottoInputDTO(numGames, numEach, fixed, mostSorted, gameType)
	input, err := NewLottoInput(dto)
	r.NoError(err)

	rgg := NewMostSortedShuffle(input)

	numbersK, numbersNK := make([]int, len(rgg.mostSortedNumbers)), make([]int, len(rgg.remainingNumbers))
	copy(numbersK, rgg.mostSortedNumbers)
	copy(numbersNK, rgg.remainingNumbers)

	numbersK, _ = pickRandomValue(numbersK)
	r.Equal(len(rgg.mostSortedNumbers)-1, len(numbersK))

	numbersNK, _ = pickRandomValue(numbersNK)
	r.Equal(len(rgg.remainingNumbers)-1, len(numbersNK))
}

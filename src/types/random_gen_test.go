package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateValidGame_RG_NoRepeatedNumbers(t *testing.T) {
	r := require.New(t)

	numGames := 100
	numEach := 13
	fixed := []int{13, 41, 60, 78}
	mostSorted := []int{5, 7, 12, 21, 25, 32, 37, 39, 45, 51, 55, 56, 61, 64, 74, 80}
	gameType := "Quina-Brasil"
	dto := newLottoInputDTO(numGames, numEach, fixed, mostSorted, gameType)
	input, err := NewLottoInput(dto)
	r.NoError(err)

	for i := 0; i < maxRepetitions; i++ {
		counter := make(map[int]bool)
		rgg := NewRandomGameGenerator(input)
		game := rgg.GenerateValidGame()

		for _, num := range game {
			if _, ok := counter[num]; ok {
				r.Failf("a game should not have repeated numbers", "got: %+v", game)
			}
			counter[num] = true
		}
	}
}

func TestGenerateLottoCombination_RG(t *testing.T) {
	r := require.New(t)

	numGames := 100
	numEach := 13
	fixed := []int{13, 41, 60, 78}
	mostSorted := []int{5, 7, 12, 21, 25, 32, 37, 39, 45, 51, 55, 56, 61, 64, 74, 80}
	gameType := "Quina-Brasil"
	dto := newLottoInputDTO(numGames, numEach, fixed, mostSorted, gameType)
	input, err := NewLottoInput(dto)
	r.NoError(err)

	rgg := NewRandomGameGenerator(input)
	lotto := rgg.GenerateLottoCombination()
	r.Equal(*dto.NumGames, lotto.Numbers.Rows)
	r.Equal(*dto.NumEachGame, lotto.Numbers.Columns)
	r.Equal(*dto.GameType, lotto.GameType)
	r.Equal(*dto.Alias, lotto.Alias)
}

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateValidGame_MS_NoRepeatedNumbers(t *testing.T) {
	r := require.New(t)
	dto := newValidLottoInputDTO()

	for i := 0; i < maxRepetitions; i++ {
		counter := make(map[int]bool)
		rgg := NewMostSortedShuffle(dto)
		game := rgg.GenerateValidGame()

		for _, num := range game {
			if _, ok := counter[num]; ok {
				r.Failf("a game should not have repeated numbers", "got: %+v", game)
			}
			counter[num] = true
		}
	}
}

func TestGenerateLottoCombination_MS(t *testing.T) {
	r := require.New(t)

	dto := newValidLottoInputDTO()
	rgg := NewMostSortedShuffle(dto)
	lotto := rgg.GenerateLottoCombination()
	r.Equal(*dto.NumGames, lotto.Numbers.Rows)
	r.Equal(*dto.NumEachGame, lotto.Numbers.Columns)
	r.Equal(*dto.GameType, lotto.GameType)
	r.Equal(*dto.Alias, lotto.Alias)
}

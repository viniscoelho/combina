package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateValidGame_RG_NoRepeatedNumbers(t *testing.T) {
	r := require.New(t)
	dto := newValidLottoInputDTO()

	for i := 0; i < maxRepetitions; i++ {
		counter := make(map[int]bool)
		rgg := NewRandomGameGenerator(dto)
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

	dto := newValidLottoInputDTO()
	rgg := NewRandomGameGenerator(dto)
	lotto := rgg.GenerateLottoCombination()
	r.Equal(*dto.NumGames, lotto.Numbers.Rows)
	r.Equal(*dto.NumEachGame, lotto.Numbers.Columns)
	r.Equal(*dto.GameType, lotto.GameType)
	r.Equal(*dto.Alias, lotto.Alias)
}

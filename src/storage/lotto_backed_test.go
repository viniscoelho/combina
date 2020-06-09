package storage

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/combina/src/types"
	"github.com/stretchr/testify/require"
)

func newLottoInputDTO(numGames, numEachGame int, fixedNumbers, mostSortedNumbers []int, gameType string) types.LottoInputDTO {
	dto := types.LottoInputDTO{
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

func TestGenerateLottoCombination_RG_SameGame(t *testing.T) {
	r := require.New(t)

	numGames := 2000
	numEach := 13
	fixed := []int{}
	mostSorted := []int{}
	gameType := "Quina-Brasil"

	dto := newLottoInputDTO(numGames, numEach, fixed, mostSorted, gameType)
	input, err := types.NewLottoInput(dto)
	r.NoError(err)

	rgg := types.NewRandomGameGenerator(input)
	lotto := rgg.GenerateLottoCombination()

	lb, err := NewFakeLottoBacked()
	r.NoError(err)

	err = lb.AddCombination(lotto)
	r.NoError(err)

	cases := make([][]int, 0)
	cases = append(cases, []int{10, 56, 62, 70, 78})
	for i := 0; i < 10; i++ {
		numbers := make([]int, 0)
		for num := 1; num <= 80; num++ {
			numbers = append(numbers, num)
		}

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(numbers), func(i, j int) {
			numbers[i], numbers[j] = numbers[j], numbers[i]
		})

		cases = append(cases, numbers[:5])
	}
	for i := 0; i < len(cases); i++ {
		r := require.New(t)
		t.Run("dummy", func(t *testing.T) {
			ans, err := lb.EvaluateCombination(lotto.ID, cases[i])
			r.NoError(err)
			log.Printf("results: %+v", ans)
		})
	}
}

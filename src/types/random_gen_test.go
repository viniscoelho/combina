package types

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/combina/src/types/ds"
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

func TestGenerateLottoCombination_RG_PossibleEntries(t *testing.T) {
	r := require.New(t)

	cases := make([]testCases, 0)
	for f := 0; f <= 4; f++ {
		for ng := 1; ng <= 5; ng++ {
			numGames := 100
			numEach := 13
			fixed := []int{13, 41, 62, 78}
			gameType := "Quina-Brasil"
			result := []int{12, 25, 51, 60, 74}

			cur := testCases{
				name:       fmt.Sprintf("%v games -- %v fixed", ng*numGames, f),
				numGames:   ng * numGames,
				numEach:    numEach,
				fixed:      fixed,
				mostSorted: []int{},
				gameType:   gameType,
				result:     result,
			}

			s := ds.NewIntStackFromSlice(cur.result)
			i := 0
			for !s.IsEmpty() && i < f {
				value, err := s.Pop()
				r.NoError(err)
				cur.fixed[i] = value
				i++
			}
			cases = append(cases, cur)
		}
	}
	for _, tc := range cases {
		r := require.New(t)
		t.Run(tc.name, func(t *testing.T) {
			dto := newLottoInputDTO(tc.numGames, tc.numEach, tc.fixed, tc.mostSorted, tc.gameType)
			input, err := NewLottoInput(dto)
			r.NoError(err)

			rgg := NewRandomGameGenerator(input)
			lotto := rgg.GenerateLottoCombination()
			ans, err := evaluateCombination(lotto, tc.result)
			r.NoError(err)
			log.Printf("num_games: %v -- results: %+v", tc.numGames, ans)
		})
	}
}

func TestGenerateLottoCombination_RG_SameGame(t *testing.T) {
	r := require.New(t)

	numGames := 250
	numEach := 13
	fixed := []int{}
	mostSorted := []int{}
	gameType := "Quina-Brasil"

	dto := newLottoInputDTO(numGames, numEach, fixed, mostSorted, gameType)
	input, err := NewLottoInput(dto)
	r.NoError(err)

	rgg := NewRandomGameGenerator(input)
	lotto := rgg.GenerateLottoCombination()

	cases := make([][]int, 0)
	for i := 0; i < 50; i++ {
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
			ans, err := evaluateCombination(lotto, cases[i])
			r.NoError(err)
			log.Printf("results: %+v", ans)
		})
	}
}

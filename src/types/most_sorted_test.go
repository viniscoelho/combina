package types

import (
	"fmt"
	"log"
	"testing"

	"github.com/combina/src/types/ds"
	"github.com/stretchr/testify/require"
)

type testCases struct {
	name       string
	numGames   int
	numEach    int
	fixed      []int
	mostSorted []int
	gameType   string
	result     []int
}

func TestGenerateValidGame_MS_NoRepeatedNumbers(t *testing.T) {
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
		rgg := NewMostSortedShuffle(input)
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

	numGames := 100
	numEach := 13
	fixed := []int{13, 41, 60, 78}
	mostSorted := []int{5, 7, 12, 21, 25, 32, 37, 39, 45, 51, 55, 56, 61, 64, 74, 80}
	gameType := "Quina-Brasil"
	dto := newLottoInputDTO(numGames, numEach, fixed, mostSorted, gameType)
	input, err := NewLottoInput(dto)
	r.NoError(err)

	rgg := NewMostSortedShuffle(input)
	lotto := rgg.GenerateLottoCombination()
	r.Equal(*dto.NumGames, lotto.Numbers.Rows)
	r.Equal(*dto.NumEachGame, lotto.Numbers.Columns)
	r.Equal(*dto.GameType, lotto.GameType)
	r.Equal(*dto.Alias, lotto.Alias)
}

func TestGenerateLottoCombination_MS_PossibleEntries(t *testing.T) {
	r := require.New(t)

	cases := make([]testCases, 0)
	for f := 0; f <= 4; f++ {
		for ms := 0; ms <= 5-f; ms++ {
			for ng := 1; ng <= 5; ng++ {
				numGames := 100
				numEach := 13
				fixed := []int{13, 41, 62, 78}
				mostSorted := []int{5, 7, 14, 21, 29, 32, 37, 39, 45, 50, 55, 56, 61, 64, 71, 80}
				gameType := "Quina-Brasil"
				result := []int{12, 25, 51, 60, 74}

				cur := testCases{
					name:       fmt.Sprintf("%v games -- %v fixed, %v most sorted", ng*numGames, f, ms),
					numGames:   ng * numGames,
					numEach:    numEach,
					fixed:      fixed,
					mostSorted: mostSorted,
					gameType:   gameType,
					result:     result,
				}

				s := ds.NewIntStackFromSlice(cur.result)
				i, j := 0, 0
				for !s.IsEmpty() && i < f {
					value, err := s.Pop()
					r.NoError(err)
					cur.fixed[i] = value
					i++
				}

				for !s.IsEmpty() && j < ms {
					value, err := s.Pop()
					r.NoError(err)
					cur.mostSorted[j] = value
					j++
				}
				cases = append(cases, cur)
			}
		}
	}
	for _, tc := range cases {
		r := require.New(t)
		t.Run(tc.name, func(t *testing.T) {
			dto := newLottoInputDTO(tc.numGames, tc.numEach, tc.fixed, tc.mostSorted, tc.gameType)
			input, err := NewLottoInput(dto)
			r.NoError(err)

			rgg := NewMostSortedShuffle(input)
			lotto := rgg.GenerateLottoCombination()
			ans, err := evaluateCombination(lotto, tc.result)
			r.NoError(err)
			log.Printf("num_games: %v -- results: %+v", tc.numGames, ans)
		})
	}
}

func evaluateCombination(l Lotto, results []int) (map[int]int, error) {
	lookup := make(map[int]bool)
	for _, r := range results {
		lookup[r] = true
	}

	ans := make(map[int]int)
	for i := 0; i < l.Numbers.Rows; i++ {
		var count int
		for j := 0; j < l.Numbers.Columns; j++ {
			num := l.Numbers.Combination[i][j]
			if _, ok := lookup[num]; ok {
				count++
			}
		}
		if count > 2 {
			ans[count]++
		}
	}

	return ans, nil
}

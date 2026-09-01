package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ---- randomGameGenerator.GenerateCombination ----

func TestRGG_GenerateCombination_CorrectLength(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	rgg := NewRandomGameGenerator(input, nil)
	combo := rgg.GenerateCombination()
	r.NotNil(combo)

	need := input.NumEachGame - len(input.FixedNumbers)
	r.Equal(need, len(combo))
}

func TestRGG_GenerateCombination_ReturnsNilWhenPoolTooSmall(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	rgg := NewRandomGameGenerator(input, nil)

	// need = 13 - 4 = 9; zero out all but 8 entries so pool < need
	need := input.NumEachGame - len(input.FixedNumbers)
	zeroed := 0
	for k := range rgg.repeated {
		if len(rgg.repeated)-zeroed <= need-1 {
			break
		}
		rgg.repeated[k] = 0
		zeroed++
	}

	combo := rgg.GenerateCombination()
	r.Nil(combo)
}

func TestRGG_GenerateCombination_AllDistinct(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	for i := 0; i < 500; i++ {
		rgg := NewRandomGameGenerator(input, nil)
		combo := rgg.GenerateCombination()
		r.NotNil(combo)

		seen := make(map[int]bool, len(combo))
		for _, n := range combo {
			r.False(seen[n], "duplicate number %d in combo %v", n, combo)
			seen[n] = true
		}
	}
}

// ---- weightedGenerator.GenerateCombination ----

func TestFY_GenerateCombination_CorrectLength(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewWeightedGenerator(input, nil)
	combo := fy.GenerateCombination()
	r.NotNil(combo)

	need := input.NumEachGame - len(input.FixedNumbers)
	r.Equal(need, len(combo))
}

func TestFY_GenerateCombination_ReturnsNilWhenPoolTooSmall(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewWeightedGenerator(input, nil)

	// need = 13 - 4 = 9; zero out all but 8 entries so pool < need
	need := input.NumEachGame - len(input.FixedNumbers)
	zeroed := 0
	for k := range fy.repeated {
		if len(fy.repeated)-zeroed <= need-1 {
			break
		}
		fy.repeated[k] = 0
		zeroed++
	}

	combo := fy.GenerateCombination()
	r.Nil(combo)
}

func TestFY_GenerateCombination_AllDistinct(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	for i := 0; i < 500; i++ {
		fy := NewWeightedGenerator(input, nil)
		combo := fy.GenerateCombination()
		r.NotNil(combo)

		seen := make(map[int]bool, len(combo))
		for _, n := range combo {
			r.False(seen[n], "duplicate number %d in combo %v", n, combo)
			seen[n] = true
		}
	}
}

func TestFY_GenerateCombination_FavoredAppearsMoreThanNeutral(t *testing.T) {
	r := require.New(t)

	// favored=[2,4,6,8,10], neutral=rest; over many games favored should appear more
	fixed := []int{13, 41, 60, 78}
	favored := []int{2, 4, 6, 8, 10}
	dto := newLottoInputDTO(200, 13, fixed, favored, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewWeightedGenerator(input, nil)
	lotto := fy.GenerateLottoCombination()

	freq := map[int]int{}
	for _, game := range lotto.Numbers.Combination {
		for _, n := range game {
			freq[n]++
		}
	}

	favoredTotal := 0
	for _, n := range favored {
		favoredTotal += freq[n]
	}
	neutralTotal := 0
	neutralCount := 0
	for n := 1; n <= 80; n++ {
		isFav := false
		for _, f := range favored {
			if f == n {
				isFav = true
				break
			}
		}
		isFixed := false
		for _, f := range fixed {
			if f == n {
				isFixed = true
				break
			}
		}
		if !isFav && !isFixed {
			neutralTotal += freq[n]
			neutralCount++
		}
	}
	favoredAvg := float64(favoredTotal) / float64(len(favored))
	neutralAvg := float64(neutralTotal) / float64(neutralCount)
	r.Greater(favoredAvg, neutralAvg, "favored avg %.1f should exceed neutral avg %.1f", favoredAvg, neutralAvg)
}

func TestFY_GenerateCombination_NeutralNumbersEvenDistribution(t *testing.T) {
	r := require.New(t)

	// no favored/disfavored — adaptive weighting should produce even spread
	dto := newLottoInputDTO(150, 15, []int{}, []int{}, "Lotofacil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewWeightedGenerator(input, nil)
	lotto := fy.GenerateLottoCombination()

	freq := map[int]int{}
	for _, game := range lotto.Numbers.Combination {
		for _, n := range game {
			freq[n]++
		}
	}

	min, max := 1<<31-1, 0
	for n := 1; n <= 25; n++ {
		if freq[n] < min {
			min = freq[n]
		}
		if freq[n] > max {
			max = freq[n]
		}
	}
	spread := max - min
	r.LessOrEqual(spread, 10, "frequency spread %d (min=%d max=%d) should be ≤10", spread, min, max)
}

// ---- weightedGenerator.isValidGame ----

func TestFY_IsValidGame_ReturnsTrueWhenAllRepeatedPositive(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewWeightedGenerator(input, nil)
	// All numbers in repeated have budget > 0 after initialize
	numbers := []int{2, 4, 6, 8, 10}
	for _, n := range numbers {
		fy.repeated[n] = 3
	}
	r.True(fy.isValidGame(numbers))
}

func TestFY_IsValidGame_ReturnsFalseWhenAnyRepeatedZeroOrNegative(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewWeightedGenerator(input, nil)
	numbers := []int{2, 4, 6, 8, 10}
	for _, n := range numbers {
		fy.repeated[n] = 3
	}
	fy.repeated[6] = 0

	r.False(fy.isValidGame(numbers))
}

func TestFY_IsValidGame_ReturnsFalseWhenRepeatedNegative(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewWeightedGenerator(input, nil)
	numbers := []int{2, 4, 6}
	for _, n := range numbers {
		fy.repeated[n] = 1
	}
	fy.repeated[4] = -1

	r.False(fy.isValidGame(numbers))
}

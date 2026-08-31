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

// ---- fisherYatesModified.GenerateCombination ----

func TestFY_GenerateCombination_CorrectLength(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewMostSortedShuffle(input, nil)
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

	fy := NewMostSortedShuffle(input, nil)

	// need = 13 - 4 = 9; shrink both pools so their combined size < 9
	fy.mostSortedNumbers = []int{1, 2, 3}
	fy.remainingNumbers = []int{4, 5}

	combo := fy.GenerateCombination()
	r.Nil(combo)
}

func TestFY_GenerateCombination_AllDistinct(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	for i := 0; i < 500; i++ {
		fy := NewMostSortedShuffle(input, nil)
		combo := fy.GenerateCombination()
		r.NotNil(combo)

		seen := make(map[int]bool, len(combo))
		for _, n := range combo {
			r.False(seen[n], "duplicate number %d in combo %v", n, combo)
			seen[n] = true
		}
	}
}

func TestFY_GenerateCombination_OnlyMostSortedWhenRemainingEmpty(t *testing.T) {
	r := require.New(t)

	// 4 fixed, need 9 picks; supply exactly 9 mostSorted and no remaining
	fixed := []int{13, 41, 60, 78}
	mostSorted := []int{2, 4, 6, 8, 10, 14, 16, 18, 20}
	dto := newLottoInputDTO(10, 13, fixed, mostSorted, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewMostSortedShuffle(input, nil)
	fy.remainingNumbers = []int{}

	msSet := make(map[int]bool, len(mostSorted))
	for _, n := range mostSorted {
		msSet[n] = true
	}

	for i := 0; i < 100; i++ {
		fy2 := NewMostSortedShuffle(input, nil)
		fy2.remainingNumbers = []int{}
		combo := fy2.GenerateCombination()
		r.NotNil(combo)
		for _, n := range combo {
			r.True(msSet[n], "number %d not in mostSorted set", n)
		}
	}
}

func TestFY_GenerateCombination_OnlyRemainingWhenMostSortedEmpty(t *testing.T) {
	r := require.New(t)

	// k=0 (no mostSorted); all picks come from remaining
	fixed := []int{13, 41, 60, 78}
	dto := newLottoInputDTO(10, 13, fixed, []int{}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewMostSortedShuffle(input, nil)
	remainSet := make(map[int]bool, len(fy.remainingNumbers))
	for _, n := range fy.remainingNumbers {
		remainSet[n] = true
	}

	for i := 0; i < 100; i++ {
		fy2 := NewMostSortedShuffle(input, nil)
		combo := fy2.GenerateCombination()
		r.NotNil(combo)
		for _, n := range combo {
			r.True(remainSet[n], "number %d not in remaining set", n)
		}
	}
}

// ---- fisherYatesModified.isValidGame ----

func TestFY_IsValidGame_ReturnsTrueWhenAllRepeatedPositive(t *testing.T) {
	r := require.New(t)

	dto := newLottoInputDTO(10, 13, []int{13, 41, 60, 78}, []int{5, 7, 12, 21}, "Quina-Brasil")
	input, err := NewLottoInput(dto)
	r.NoError(err)

	fy := NewMostSortedShuffle(input, nil)
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

	fy := NewMostSortedShuffle(input, nil)
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

	fy := NewMostSortedShuffle(input, nil)
	numbers := []int{2, 4, 6}
	for _, n := range numbers {
		fy.repeated[n] = 1
	}
	fy.repeated[4] = -1

	r.False(fy.isValidGame(numbers))
}

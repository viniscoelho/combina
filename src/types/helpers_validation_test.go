package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsNumEachValid(t *testing.T) {
	cases := []struct {
		name     string
		num      int
		gameType string
		want     bool
	}{
		{"Lotofacil min boundary", 15, "Lotofacil", true},
		{"Lotofacil max boundary", 18, "Lotofacil", true},
		{"Lotofacil mid", 16, "Lotofacil", true},
		{"Lotofacil below min", 14, "Lotofacil", false},
		{"Lotofacil above max", 19, "Lotofacil", false},

		{"Lotomania exact", 50, "Lotomania", true},
		{"Lotomania below", 49, "Lotomania", false},
		{"Lotomania above", 51, "Lotomania", false},

		{"Quina min boundary", 5, "Quina", true},
		{"Quina max boundary", 15, "Quina", true},
		{"Quina mid", 10, "Quina", true},
		{"Quina below min", 4, "Quina", false},
		{"Quina above max", 16, "Quina", false},

		{"Mega-Sena min boundary", 6, "Mega-Sena", true},
		{"Mega-Sena max boundary", 15, "Mega-Sena", true},
		{"Mega-Sena mid", 10, "Mega-Sena", true},
		{"Mega-Sena below min", 5, "Mega-Sena", false},
		{"Mega-Sena above max", 16, "Mega-Sena", false},

		{"Quina-Brasil exact", 13, "Quina-Brasil", true},
		{"Quina-Brasil below", 12, "Quina-Brasil", false},
		{"Quina-Brasil above", 14, "Quina-Brasil", false},

		{"Quininha min boundary", 15, "Quininha", true},
		{"Quininha max of range", 20, "Quininha", true},
		{"Quininha mid", 17, "Quininha", true},
		{"Quininha 25", 25, "Quininha", true},
		{"Quininha 30", 30, "Quininha", true},
		{"Quininha below min", 14, "Quininha", false},
		{"Quininha gap 21", 21, "Quininha", false},
		{"Quininha gap 26", 26, "Quininha", false},
		{"Quininha above 30", 31, "Quininha", false},

		{"Seninha min boundary", 14, "Seninha", true},
		{"Seninha max of range", 20, "Seninha", true},
		{"Seninha mid", 17, "Seninha", true},
		{"Seninha 25", 25, "Seninha", true},
		{"Seninha 30", 30, "Seninha", true},
		{"Seninha below min", 13, "Seninha", false},
		{"Seninha gap 21", 21, "Seninha", false},
		{"Seninha gap 26", 26, "Seninha", false},
		{"Seninha above 30", 31, "Seninha", false},

		{"unknown game type", 10, "Unknown", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsNumEachValid(tc.num, tc.gameType))
		})
	}
}

func TestIsValidNumbers(t *testing.T) {
	cases := []struct {
		name    string
		r       MinMaxRange
		numbers []int
		want    bool
	}{
		{"empty slice always valid", MinMaxRange{1, 60}, []int{}, true},
		{"all within range", MinMaxRange{1, 60}, []int{1, 30, 60}, true},
		{"single at min", MinMaxRange{1, 60}, []int{1}, true},
		{"single at max", MinMaxRange{1, 60}, []int{60}, true},
		{"one below min", MinMaxRange{1, 60}, []int{0, 30, 60}, false},
		{"one above max", MinMaxRange{1, 60}, []int{1, 30, 61}, false},
		{"all below min", MinMaxRange{1, 60}, []int{-1, -5}, false},
		{"all above max", MinMaxRange{1, 60}, []int{61, 99}, false},
		{"Lotomania allows 0", MinMaxRange{0, 99}, []int{0, 50, 99}, true},
		{"Lotomania below 0", MinMaxRange{0, 99}, []int{-1}, false},
		{"Lotomania above 99", MinMaxRange{0, 99}, []int{100}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isValidNumbers(tc.r, tc.numbers))
		})
	}
}

func TestIsValidNumGames(t *testing.T) {
	cases := []struct {
		name       string
		numGames   int64
		maxRange   int
		numEach    int
		numFixed   int
		want       bool
	}{
		// r > 20 → always true (Lotomania path)
		{"Lotomania r=50 always true", 1_000_000, 99, 50, 0, true},
		{"r=21 always true", 1_000_000, 100, 21, 0, true},

		// small combinatorial space: C(5,3)=10
		{"C(5,3) numGames=1 valid", 1, 5, 3, 0, true},
		{"C(5,3) numGames=10 at boundary valid", 10, 5, 3, 0, true},
		{"C(5,3) numGames=11 exceeds invalid", 11, 5, 3, 0, false},

		// C(6,3)=20
		{"C(6,3) numGames=20 at boundary valid", 20, 6, 3, 0, true},
		{"C(6,3) numGames=21 exceeds invalid", 21, 6, 3, 0, false},

		// Quina-Brasil: C(80,13)=1,646,492,110,120 — large but valid
		{"Quina-Brasil numGames=2000 valid", 2000, 80, 13, 0, true},
		{"Quina-Brasil numGames=1 valid", 1, 80, 13, 0, true},

		// with fixed numbers: C(n-numFixed, r-numFixed)
		// numFixed=1 → n=4, r=2, C(4,2)=6
		{"C(5,3) numFixed=1 → C(4,2)=6 valid", 6, 5, 3, 1, true},
		{"C(5,3) numFixed=1 → C(4,2)=6 invalid", 7, 5, 3, 1, false},

		// Mega-Sena: C(60,6)=50,063,860
		{"Mega-Sena numGames=50000000 valid", 50_000_000, 60, 6, 0, true},
		{"Mega-Sena numGames=50063860 at boundary valid", 50_063_860, 60, 6, 0, true},
		{"Mega-Sena numGames=50063861 exceeds invalid", 50_063_861, 60, 6, 0, false},

		// Lotofacil: C(25,15)=3,268,760
		{"Lotofacil numGames=3268760 at boundary valid", 3_268_760, 25, 15, 0, true},
		{"Lotofacil numGames=3268761 exceeds invalid", 3_268_761, 25, 15, 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isValidNumGames(tc.numGames, tc.maxRange, tc.numEach, tc.numFixed))
		})
	}
}

func TestValidateIntersection(t *testing.T) {
	cases := []struct {
		name       string
		fixed      []int
		excluded   []int
		favored    []int
		disfavored []int
		wantErr    bool
	}{
		// all empty
		{"all empty no error", nil, nil, nil, nil, false},

		// no overlap across all four lists
		{"all four no overlap", []int{1, 2}, []int{3, 4}, []int{5, 6}, []int{7, 8}, false},

		// fixed vs others
		{"fixed vs excluded", []int{1, 2}, []int{2, 3}, nil, nil, true},
		{"fixed vs favored", []int{1, 5}, []int{}, []int{5, 10}, nil, true},
		{"fixed vs disfavored", []int{1, 5}, nil, nil, []int{5, 10}, true},

		// excluded vs others
		{"excluded vs favored", nil, []int{3, 7}, []int{7, 9}, nil, true},
		{"excluded vs disfavored", nil, []int{3, 7}, nil, []int{7, 9}, true},

		// favored vs disfavored
		{"favored vs disfavored", nil, nil, []int{1, 2, 3}, []int{3, 4, 5}, true},

		// single-element lists
		{"single element no overlap", []int{1}, []int{2}, []int{3}, []int{4}, false},
		{"single element overlap fixed excluded", []int{5}, []int{5}, nil, nil, true},

		// partial population
		{"only fixed and favored no overlap", []int{1, 2, 3}, nil, []int{4, 5, 6}, nil, false},
		{"only favored and disfavored no overlap", nil, nil, []int{1, 2}, []int{3, 4}, false},
		{"only excluded no error", nil, []int{10, 20}, nil, nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWeightedIntersection(tc.fixed, tc.excluded, tc.favored, tc.disfavored)
			if tc.wantErr {
				require.Error(t, err)
				var target InvalidDTOError
				require.ErrorAs(t, err, &target)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

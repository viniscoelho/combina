//go:generate mockgen -destination=mocks/mocks.go -package=mocks combina/src/types LottoCombinator,RandomGameGenerator
package types

import (
	"time"

	_ "github.com/golang/mock/mockgen/model"
)

type MinMaxRange struct {
	Min int
	Max int
}

type LottoInput struct {
	NumGames          int
	NumEachGame       int
	FixedNumbers      []int
	MostSortedNumbers []int
	GameType          string
	Alias             string
}

type LottoInputDTO struct {
	NumGames          *int    `json:"num_games"`
	NumEachGame       *int    `json:"num_each"`
	FixedNumbers      []int   `json:"fixed_numbers"`
	MostSortedNumbers []int   `json:"most_sorted"`
	GameType          *string `json:"game_type"`
	Alias             *string `json:"alias,omitempty"`
}

type Lotto struct {
	ID        string    `json:"id"`
	Numbers   GameCombo `json:"numbers"`
	GameType  string    `json:"game_type"`
	CreatedOn time.Time `json:"created_on"`
	Alias     string    `json:"alias"`
}

type GameCombo struct {
	Combination [][]int `json:"combination"`
	Rows        int     `json:"rows"`
	Columns     int     `json:"cols"`
}

type LottoCombinator interface {
	ListCombinations(gameType string) ([]Lotto, error)
	AddCombination(lotto Lotto) error
	FetchCombination(id string) (Lotto, error)
	DeleteCombination(id string) error
	EvaluateCombination(id string, results []int) (map[int]int, error)
}

type RandomGameGenerator interface {
	GenerateCombination() []int
	GenerateValidGame() []int
	GenerateLottoCombination() Lotto
}

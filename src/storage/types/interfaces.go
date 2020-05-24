//go:generate mockgen -destination=mocks/mocks.go -package=mocks combina/src/types Combination
package types

import "time"

type LottoInputDTO struct {
	NumGames     *int    `json:"num_games"`
	NumEachGame  *int    `json:"num_each"`
	FixedNumbers []int   `json:"fixed_numbers"`
	GameType     *string `json:"game_type"`
	Alias        *string `json:"alias"`
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

type Combination interface {
	ListCombinations(gameType string) ([]Lotto, error)
	CreateCombination(lotto Lotto) error
	ReadCombination(id string) (Lotto, error)
	DeleteCombination(id string) error
}

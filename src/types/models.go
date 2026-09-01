package types

import "time"

type MinMaxRange struct {
	Min int
	Max int
}

type LottoInput struct {
	NumGames          int
	NumEachGame       int
	FixedNumbers      []int
	FavoredNumbers []int
	DisfavoredNumbers []int
	ExcludedNumbers   []int
	GameType          string
	Alias             string
}

type LottoInputDTO struct {
	NumGames          *int    `json:"num_games"`
	NumEachGame       *int    `json:"num_each"`
	FixedNumbers      []int   `json:"fixed_numbers"`
	FavoredNumbers    []int   `json:"favored,omitempty"`
	DisfavoredNumbers []int   `json:"disfavored,omitempty"`
	ExcludedNumbers   []int   `json:"excluded_numbers,omitempty"`
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

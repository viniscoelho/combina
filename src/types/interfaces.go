//go:generate mockgen -destination=mocks/mocks.go -package=mocks combina/src/types LottoCombinator,GameGenerator
package types

import (
	_ "github.com/golang/mock/mockgen/model"
)

type LottoCombinator interface {
	ListCombinations(gameType string) ([]Lotto, error)
	AddCombination(lotto Lotto) error
	UpdateCombination(lotto Lotto) error
	FetchCombination(id string) (Lotto, error)
	DeleteCombination(id string) error
	EvaluateCombination(id string, results []int) (map[int]int, error)
}

type GameGenerator interface {
	GenerateCombination() []int
	GenerateValidGame() []int
	GenerateLottoCombination() Lotto
}

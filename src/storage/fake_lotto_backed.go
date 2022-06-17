package storage

import (
	"sort"

	"combina/src/types"
)

type fakeLottoBacked struct {
	storage map[string]types.Lotto
}

func NewFakeLottoBacked() (*fakeLottoBacked, error) {
	lb := &fakeLottoBacked{
		storage: make(map[string]types.Lotto, 0),
	}

	return lb, nil
}

func (lb fakeLottoBacked) ListCombinations(gameType string) ([]types.Lotto, error) {
	if _, ok := types.Games[gameType]; !ok && gameType != "" {
		return nil, types.GameTypeDoesNotExistError{}
	}

	ll := make([]types.Lotto, 0)
	for gc, lotto := range lb.storage {
		if gc != "" && gc == gameType {
			ll = append(ll, lotto)
		} else {
			ll = append(ll, lotto)
		}
	}

	// sort by desc order
	sort.SliceStable(ll, func(i, j int) bool {
		return ll[i].CreatedOn.After(ll[j].CreatedOn)
	})

	return ll, nil
}

func (lb *fakeLottoBacked) AddCombination(lotto types.Lotto) error {
	if _, ok := lb.storage[lotto.ID]; ok {
		return types.CombinationAlreadyExistsError{}
	}

	lb.storage[lotto.ID] = lotto
	return nil
}

func (lb fakeLottoBacked) FetchCombination(id string) (types.Lotto, error) {
	l, ok := lb.storage[id]
	if !ok {
		return types.Lotto{}, types.CombinationDoesNotExistError{}
	}

	return l, nil
}

func (lb *fakeLottoBacked) DeleteCombination(id string) error {
	if _, ok := lb.storage[id]; !ok {
		return types.CombinationDoesNotExistError{}
	}

	delete(lb.storage, id)
	return nil
}

func (lb fakeLottoBacked) EvaluateCombination(id string, result []int) (map[int]int, error) {
	l, ok := lb.storage[id]
	if !ok {
		return nil, types.CombinationDoesNotExistError{}
	}

	lookup := make(map[int]struct{})
	for _, r := range result {
		lookup[r] = struct{}{}
	}

	scores := make(map[int]int)
	for i := 0; i < l.Numbers.Rows; i++ {
		var count int
		for j := 0; j < l.Numbers.Columns; j++ {
			num := l.Numbers.Combination[i][j]
			if _, ok := lookup[num]; ok {
				count++
			}
		}
		scores[count]++
	}

	return lb.filterResults(l.GameType, scores), nil
}

func (lb fakeLottoBacked) filterResults(gameType string, scores map[int]int) map[int]int {
	filtered := make(map[int]int)
	for _, p := range types.Prizes[gameType] {
		if _, ok := scores[p]; ok {
			filtered[p] = scores[p]
		}
	}

	return filtered
}

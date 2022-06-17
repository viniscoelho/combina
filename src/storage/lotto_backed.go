package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"combina/src/db"
	"combina/src/types"
	"github.com/jackc/pgx/v4"
)

type lottoBacked struct {
	storage map[string]types.Lotto
}

func NewLottoBacked() (*lottoBacked, error) {
	lb := &lottoBacked{
		storage: make(map[string]types.Lotto, 0),
	}

	err := lb.initializeLottoBacked()
	if err != nil {
		return nil, err
	}

	return lb, nil
}

func (lb *lottoBacked) initializeLottoBacked() error {
	conn, err := db.DatabaseConnect(types.DatabaseName)
	if err != nil {
		return err
	}
	defer conn.Close()

	rows, err := conn.Query(context.Background(), "SELECT * FROM lotto")
	if err != nil {
		return err
	}

	for rows.Next() {
		lotto := types.Lotto{}
		if err = rows.Scan(&lotto.ID, &lotto.GameType, &lotto.Numbers, &lotto.CreatedOn, &lotto.Alias); err != nil {
			return err
		}

		lb.storage[lotto.ID] = lotto
	}

	return nil
}

func (lb lottoBacked) ListCombinations(gameType string) ([]types.Lotto, error) {
	if _, ok := types.Games[gameType]; !ok && gameType != "" {
		return nil, types.GameTypeDoesNotExistError{}
	}

	conn, err := db.DatabaseConnect(types.DatabaseName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var rows pgx.Rows
	if gameType != "" {
		rows, err = conn.Query(context.Background(), "SELECT * FROM lotto WHERE	type = $1 ORDER BY created_on DESC", gameType)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = conn.Query(context.Background(), "SELECT * FROM lotto ORDER BY created_on DESC")
		if err != nil {
			return nil, err
		}
	}

	ll := make([]types.Lotto, 0)
	for rows.Next() {
		lotto := types.Lotto{}
		if err = rows.Scan(&lotto.ID, &lotto.GameType, &lotto.Numbers, &lotto.CreatedOn, &lotto.Alias); err != nil {
			return nil, err
		}

		ll = append(ll, lotto)
	}

	return ll, nil
}

func (lb *lottoBacked) AddCombination(lotto types.Lotto) error {
	if _, ok := lb.storage[lotto.ID]; ok {
		return types.CombinationAlreadyExistsError{}
	}

	bytes, err := json.Marshal(lotto.Numbers)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	conn, err := db.DatabaseConnect(types.DatabaseName)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(context.Background(), "INSERT INTO lotto (id, type, combination, created_on, name) VALUES ($1, $2, $3, $4, $5)",
		lotto.ID, lotto.GameType, bytes, lotto.CreatedOn, lotto.Alias)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	lb.storage[lotto.ID] = lotto
	return nil
}

func (lb lottoBacked) FetchCombination(id string) (types.Lotto, error) {
	l, ok := lb.storage[id]
	if !ok {
		return types.Lotto{}, types.CombinationDoesNotExistError{}
	}

	return l, nil
}

func (lb *lottoBacked) DeleteCombination(id string) error {
	if _, ok := lb.storage[id]; !ok {
		return types.CombinationDoesNotExistError{}
	}

	conn, err := db.DatabaseConnect(types.DatabaseName)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(context.Background(), "DELETE FROM lotto WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deletion failed: %w", err)
	}

	delete(lb.storage, id)
	return nil
}

func (lb lottoBacked) EvaluateCombination(id string, result []int) (map[int]int, error) {
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

func (lb lottoBacked) filterResults(gameType string, scores map[int]int) map[int]int {
	filtered := make(map[int]int)
	for _, p := range types.Prizes[gameType] {
		if _, ok := scores[p]; ok {
			filtered[p] = scores[p]
		}
	}

	return filtered
}

package storage

import (
	"testing"
	"time"

	"combina/src/types"
	"github.com/stretchr/testify/require"
)

func newLotto(id, gameType string, createdOn time.Time, combo [][]int) types.Lotto {
	return types.Lotto{
		ID:       id,
		GameType: gameType,
		CreatedOn: createdOn,
		Numbers: types.GameCombo{
			Combination: combo,
			Rows:        len(combo),
			Columns:     func() int {
				if len(combo) > 0 {
					return len(combo[0])
				}
				return 0
			}(),
		},
	}
}

func newBackend(t *testing.T) *fakeLottoBacked {
	t.Helper()
	lb, err := NewFakeLottoBacked()
	require.NoError(t, err)
	return lb
}

func TestListCombinations_Empty(t *testing.T) {
	lb := newBackend(t)
	results, err := lb.ListCombinations("")
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestListCombinations_AllGames(t *testing.T) {
	lb := newBackend(t)
	base := time.Now()
	require.NoError(t, lb.AddCombination(newLotto("a1", "Mega-Sena", base, [][]int{{1, 2, 3, 4, 5, 6}})))
	require.NoError(t, lb.AddCombination(newLotto("b1", "Quina", base.Add(time.Second), [][]int{{1, 2, 3, 4, 5}})))

	results, err := lb.ListCombinations("")
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestListCombinations_FilterByGameType(t *testing.T) {
	lb := newBackend(t)
	base := time.Now()
	require.NoError(t, lb.AddCombination(newLotto("a1", "Mega-Sena", base, [][]int{{1, 2, 3, 4, 5, 6}})))
	require.NoError(t, lb.AddCombination(newLotto("b1", "Quina", base.Add(time.Second), [][]int{{1, 2, 3, 4, 5}})))

	results, err := lb.ListCombinations("Mega-Sena")
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestListCombinations_InvalidGameType(t *testing.T) {
	lb := newBackend(t)
	_, err := lb.ListCombinations("NotAGame")
	require.ErrorAs(t, err, &types.GameTypeDoesNotExistError{})
}

func TestListCombinations_SortedDescByCreatedOn(t *testing.T) {
	lb := newBackend(t)
	base := time.Now()
	older := newLotto("old", "Quina", base, [][]int{{1, 2, 3, 4, 5}})
	newer := newLotto("new", "Quina", base.Add(time.Hour), [][]int{{6, 7, 8, 9, 10}})
	require.NoError(t, lb.AddCombination(older))
	require.NoError(t, lb.AddCombination(newer))

	results, err := lb.ListCombinations("")
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.True(t, results[0].CreatedOn.After(results[1].CreatedOn))
}

func TestAddCombination_Success(t *testing.T) {
	lb := newBackend(t)
	l := newLotto("x1", "Quina", time.Now(), [][]int{{1, 2, 3, 4, 5}})
	require.NoError(t, lb.AddCombination(l))
}

func TestAddCombination_Duplicate(t *testing.T) {
	lb := newBackend(t)
	l := newLotto("x1", "Quina", time.Now(), [][]int{{1, 2, 3, 4, 5}})
	require.NoError(t, lb.AddCombination(l))
	err := lb.AddCombination(l)
	require.ErrorAs(t, err, &types.CombinationAlreadyExistsError{})
}

func TestFetchCombination_Found(t *testing.T) {
	lb := newBackend(t)
	l := newLotto("x1", "Quina", time.Now(), [][]int{{1, 2, 3, 4, 5}})
	require.NoError(t, lb.AddCombination(l))
	got, err := lb.FetchCombination("x1")
	require.NoError(t, err)
	require.Equal(t, l, got)
}

func TestFetchCombination_NotFound(t *testing.T) {
	lb := newBackend(t)
	_, err := lb.FetchCombination("missing")
	require.ErrorAs(t, err, &types.CombinationDoesNotExistError{})
}

func TestDeleteCombination_Success(t *testing.T) {
	lb := newBackend(t)
	l := newLotto("x1", "Quina", time.Now(), [][]int{{1, 2, 3, 4, 5}})
	require.NoError(t, lb.AddCombination(l))
	require.NoError(t, lb.DeleteCombination("x1"))
	_, err := lb.FetchCombination("x1")
	require.ErrorAs(t, err, &types.CombinationDoesNotExistError{})
}

func TestDeleteCombination_NotFound(t *testing.T) {
	lb := newBackend(t)
	err := lb.DeleteCombination("missing")
	require.ErrorAs(t, err, &types.CombinationDoesNotExistError{})
}

func TestEvaluateCombination_NotFound(t *testing.T) {
	lb := newBackend(t)
	_, err := lb.EvaluateCombination("missing", []int{1, 2, 3, 4, 5, 6})
	require.ErrorAs(t, err, &types.CombinationDoesNotExistError{})
}

func TestEvaluateCombination_HitCounts(t *testing.T) {
	lb := newBackend(t)
	combo := [][]int{
		{4, 5, 6, 7, 8, 9},
		{1, 2, 3, 40, 50, 60},
	}
	l := newLotto("e1", "Mega-Sena", time.Now(), combo)
	require.NoError(t, lb.AddCombination(l))

	result := []int{4, 5, 6, 7, 8, 9}
	scores, err := lb.EvaluateCombination("e1", result)
	require.NoError(t, err)
	// first row: 6 hits → prize tier for Mega-Sena
	require.Equal(t, 1, scores[6])
	// second row: 0 hits → not in Prizes["Mega-Sena"] = {4,5,6}, not returned
	require.NotContains(t, scores, 0)
}

func TestEvaluateCombination_FilteredByPrizes(t *testing.T) {
	lb := newBackend(t)
	combo := [][]int{
		{1, 2, 3, 4, 5, 6},
		{10, 20, 30, 40, 50, 60},
	}
	l := newLotto("e2", "Mega-Sena", time.Now(), combo)
	require.NoError(t, lb.AddCombination(l))

	// drawn numbers match 3 from row1 and 0 from row2 — neither qualifies for Mega-Sena prizes {4,5,6}
	result := []int{1, 2, 3, 99, 98, 97}
	scores, err := lb.EvaluateCombination("e2", result)
	require.NoError(t, err)
	require.Empty(t, scores)
}

func TestEvaluateCombination_PartialHits(t *testing.T) {
	lb := newBackend(t)
	combo := [][]int{
		{1, 2, 3, 4, 5, 6},
		{1, 2, 3, 4, 5, 7},
		{10, 20, 30, 40, 50, 60},
	}
	l := newLotto("e3", "Mega-Sena", time.Now(), combo)
	require.NoError(t, lb.AddCombination(l))

	// rows 0 and 1 each have 5 matching numbers
	result := []int{1, 2, 3, 4, 5, 99}
	scores, err := lb.EvaluateCombination("e3", result)
	require.NoError(t, err)
	// 5 hits qualifies for Mega-Sena prize tier 5; two rows scored 5
	require.Equal(t, 2, scores[5])
	// row2 has 0 hits, not a prize tier
	require.NotContains(t, scores, 0)
}

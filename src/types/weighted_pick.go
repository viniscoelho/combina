package types

import "math/rand"

// weightedSample draws `need` distinct numbers from candidates without
// replacement. Each candidate's selection probability at each step is
// proportional to its integer weight (higher = more likely).
//
// Algorithm: roulette-wheel selection. At each step, compute the total weight
// of remaining candidates, draw a uniform random value in [0, total), then
// walk the pool subtracting weights until the running sum goes negative — the
// number that tipped it over is selected and swapped out of the pool.
// Repeating `need` times gives a weighted sample without replacement in O(n²)
// where n = len(candidates). This is acceptable because n ≤ ~80 (max lottery
// range) and need ≤ ~50.
//
// Assumes len(candidates) >= need and all weights > 0.
func weightedSample(candidates []int, weight func(int) int, need int) []int {
	pool := make([]int, len(candidates))
	copy(pool, candidates)
	result := make([]int, 0, need)

	for len(result) < need {
		total := 0
		for _, n := range pool {
			total += weight(n)
		}

		r := rand.Intn(total)
		for i, n := range pool {
			r -= weight(n)
			if r < 0 {
				result = append(result, n)
				pool[i] = pool[len(pool)-1]
				pool = pool[:len(pool)-1]
				break
			}
		}
	}

	return result
}

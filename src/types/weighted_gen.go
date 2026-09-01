package types

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

type weightedGenerator struct {
	// Rejects games that duplicate or are subsets of already-seen games
	dedup *dedupIndex
	// A map to count how many times a number has been used (decremented each pick)
	repeated map[int]int
	// Per-number class factor: favored=1.75, disfavored=0.5, neutral=1.0
	classFactor map[int]float64
	// Counts how many times each number has been picked across all generated games
	picked map[int]int
	// A slice having the fixed numbers
	fixedNumbers []int
	// A slice having the favored numbers (boosted probability, 1.75x)
	favoredNumbers []int
	// A slice having the disfavored numbers (reduced probability, 0.5x)
	disfavoredNumbers []int
	// A slice having numbers that must never appear in any game
	excludedNumbers []int
	// Number of games (a.k.a. combinations) to be generated
	numGames int
	// Amount of numbers for each game
	numEachGame int
	// Maximum value allowed for each non-fixed number to be used
	maxUsage int
	// Game set kind, e.g., Quina, Lotofacil, ...
	gameType string
	// Minimum and maximum allowed values for a game, e.g, [1, 80]
	gameRange MinMaxRange
	// An alias for the game set
	alias string
}

// NewWeightedGenerator creates a generator that applies per-number probability
// weights: favored numbers get 1.75× base weight, disfavored get 0.5×, normal
// numbers get 1×. Within each class, selection is adaptive: numbers picked less
// often so far are preferred, so the full set converges to an even distribution.
// Routing selects this generator when either favored or disfavored is non-empty.
//
// existing pre-seeds the duplicate-check map so that games already known
// (e.g. imported from a file) are never re-generated.
func NewWeightedGenerator(input LottoInput, existing [][]int) *weightedGenerator {
	g := weightedGenerator{}

	g.repeated = make(map[int]int)
	g.classFactor = make(map[int]float64)
	g.picked = make(map[int]int)
	g.fixedNumbers = make([]int, len(input.FixedNumbers))
	g.favoredNumbers = make([]int, len(input.FavoredNumbers))
	g.disfavoredNumbers = make([]int, len(input.DisfavoredNumbers))
	g.excludedNumbers = make([]int, len(input.ExcludedNumbers))

	copy(g.fixedNumbers, input.FixedNumbers)
	copy(g.favoredNumbers, input.FavoredNumbers)
	copy(g.disfavoredNumbers, input.DisfavoredNumbers)
	copy(g.excludedNumbers, input.ExcludedNumbers)

	g.numGames = input.NumGames
	g.numEachGame = input.NumEachGame
	g.gameRange = Games[input.GameType]
	numFixed := len(g.fixedNumbers)
	numExcluded := len(g.excludedNumbers)
	maxRange := g.gameRange.Max

	poolSize := maxRange - numFixed - numExcluded
	g.maxUsage = ((g.numEachGame-numFixed)*g.numGames)/poolSize + 1
	if ((g.numEachGame-numFixed)*g.numGames)%poolSize != 0 {
		g.maxUsage++
	}

	g.gameType = input.GameType
	g.alias = input.Alias

	g.initialize()

	g.dedup = newDedupIndex(existing, g.numEachGame)

	return &g
}

// initialize (re-)populates repeated, classFactor, and resets picked.
// Favored numbers receive 1.75× budget and classFactor, disfavored receive 0.5×,
// and regular numbers receive exactly 1×. Fixed and excluded numbers are omitted.
//
// picked is reset so the adaptive weights start from a clean slate each cycle —
// without the reset, all numbers would converge to the same pick count and the
// adaptive weighting would flatten to near-uniform, losing the favored/disfavored
// effect on the next cycle.
//
// Called once at construction and again if the pool is exhausted mid-generation.
func (g *weightedGenerator) initialize() {
	fixed := make(map[int]bool)
	favored := make(map[int]bool)
	disfavored := make(map[int]bool)
	excluded := make(map[int]bool)

	for _, num := range g.fixedNumbers {
		fixed[num] = true
	}
	for _, num := range g.favoredNumbers {
		favored[num] = true
	}
	for _, num := range g.disfavoredNumbers {
		disfavored[num] = true
	}
	for _, num := range g.excludedNumbers {
		excluded[num] = true
	}

	minRange, maxRange := g.gameRange.Min, g.gameRange.Max
	g.picked = make(map[int]int)

	for num := minRange; num <= maxRange; num++ {
		if fixed[num] || excluded[num] {
			continue
		}
		if favored[num] {
			g.repeated[num] = int(float64(g.maxUsage) * 1.75)
			g.classFactor[num] = 1.75
		} else if disfavored[num] {
			g.repeated[num] = int(float64(g.maxUsage) * 0.5)
			g.classFactor[num] = 0.5
		} else {
			g.repeated[num] = g.maxUsage
			g.classFactor[num] = 1.0
		}
	}
}

// GenerateCombination builds a candidate game using adaptive inverse-frequency
// weighting combined with per-class bias factors.
//
// For each eligible number n (not fixed, not excluded, budget > 0):
//
//	base     = maxPicked - picked[n] + 1
//	weight   = int(base × classFactor[n] × 100)
//
// where maxPicked is the highest pick count among all eligible numbers this
// call, and classFactor is 1.75 for favored, 0.5 for disfavored, 1.0 for
// neutral.
//
// The adaptive base ensures that a number already picked many times gets a
// lower base weight than a number picked rarely, steering future picks toward
// under-used numbers. This keeps neutral numbers balanced: in a batch of 150
// Lotofácil games the most and least frequent neutral numbers differ by ~3
// appearances rather than ~11 with plain uniform selection.
//
// Favored and disfavored numbers intentionally break from that balance —
// favored numbers appear roughly 1.75× more than neutral, disfavored roughly
// 0.5×. The gap between the most and least frequent number across the whole
// batch will therefore be large when favored/disfavored are set; that is
// expected and desired.
//
// The ×100 scaling converts the float classFactor to an integer weight without
// losing meaningful precision for the roulette-wheel sampler in weightedSample.
//
// Returns nil if fewer eligible numbers remain than needed to fill a game.
func (g *weightedGenerator) GenerateCombination() []int {
	need := g.numEachGame - len(g.fixedNumbers)

	candidates := make([]int, 0)
	for num, budget := range g.repeated {
		if budget > 0 {
			candidates = append(candidates, num)
		}
	}

	if len(candidates) < need {
		return nil
	}

	maxPicked := 0
	for _, n := range candidates {
		if g.picked[n] > maxPicked {
			maxPicked = g.picked[n]
		}
	}

	weight := func(n int) int {
		base := maxPicked - g.picked[n] + 1
		return int(float64(base) * g.classFactor[n] * 100)
	}

	return weightedSample(candidates, weight, need)
}

// isValidGame checks that every number in the candidate combination still has
// remaining usage budget (repeated > 0). A combination that would exceed the
// budget of any number is rejected so that no single number dominates the
// full set of generated games.
func (g *weightedGenerator) isValidGame(numbers []int) bool {
	for _, num := range numbers {
		if c := g.repeated[num]; c <= 0 {
			return false
		}
	}
	return true
}

// GenerateValidGame produces one unique game by repeatedly calling
// GenerateCombination until a result passes isValidGame and has not been
// generated before. It then appends fixed numbers, sorts, records the game in
// the duplicate-check map, and decrements usage counters for non-fixed numbers.
//
// If GenerateCombination returns nil (pool exhausted), initialize() resets the
// usage budgets before retrying, preventing an infinite loop when numGames is
// close to the combinatorial maximum.
func (g *weightedGenerator) GenerateValidGame() []int {
	fixed := make(map[int]bool, len(g.fixedNumbers))
	for _, num := range g.fixedNumbers {
		fixed[num] = true
	}

	var numbers []int
	for {
		numbers = g.GenerateCombination()
		if numbers == nil || !g.isValidGame(numbers) {
			g.initialize()
			continue
		}

		// add the fixed numbers to the result
		for _, num := range g.fixedNumbers {
			numbers = append(numbers, num)
		}
		sort.Slice(numbers, func(i, j int) bool {
			return numbers[i] < numbers[j]
		})

		if g.dedup.isDuplicate(numbers) {
			continue
		}
		g.dedup.record(numbers)

		// Decrement usage counters and increment picked only for non-fixed numbers.
		for _, num := range numbers {
			if !fixed[num] {
				g.repeated[num]--
				g.picked[num]++
			}
		}
		break
	}

	return numbers
}

// GenerateLottoCombination generates all numGames combinations and packages
// them into a Lotto value with a new UUID and the current timestamp.
func (g *weightedGenerator) GenerateLottoCombination() Lotto {
	combination := make([][]int, 0)
	for i := 0; i < g.numGames; i++ {
		numbers := g.GenerateValidGame()
		combination = append(combination, numbers)
	}

	id := uuid.New()
	gc := GameCombo{
		Combination: combination,
		Rows:        g.numGames,
		Columns:     g.numEachGame,
	}

	return Lotto{
		ID:        id.String(),
		Numbers:   gc,
		GameType:  g.gameType,
		CreatedOn: time.Now(),
		Alias:     g.alias,
	}
}

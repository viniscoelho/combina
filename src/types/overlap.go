package types

import "sort"

// OverlapRelation describes how set A relates to set B.
type OverlapRelation string

const (
	RelationEqual    OverlapRelation = "equal"
	RelationSubset   OverlapRelation = "subset"   // A ⊂ B
	RelationSuperset OverlapRelation = "superset" // A ⊃ B
	RelationPartial  OverlapRelation = "partial"  // some numbers in common
)

// OverlapFinding records an overlap between two sets within a game.
type OverlapFinding struct {
	SetA           int             `json:"set_a"`           // 0-based index
	SetB           int             `json:"set_b"`           // 0-based index
	Relation       OverlapRelation `json:"relation"`
	SharedNumbers  []int           `json:"shared_numbers"`
	SharedCount    int             `json:"shared_count"`
}

// OverlapReport is the result of analysing all sets in a game.
type OverlapReport struct {
	Findings []OverlapFinding `json:"findings"`
	Clean    bool             `json:"clean"` // true when no overlaps found
}

// AnalyseOverlaps performs a pairwise overlap check over all sets.
func AnalyseOverlaps(sets [][]int) OverlapReport {
	var findings []OverlapFinding

	for i := 0; i < len(sets); i++ {
		setI := toSet(sets[i])
		for j := i + 1; j < len(sets); j++ {
			setJ := toSet(sets[j])
			shared := sharedNumbers(setI, sets[j])
			smaller := len(sets[i])
			if len(sets[j]) < smaller {
				smaller = len(sets[j])
			}
			if len(shared)*10 < smaller*8 {
				continue
			}
			rel := classify(sets[i], sets[j], setI, setJ, shared)
			findings = append(findings, OverlapFinding{
				SetA:          i,
				SetB:          j,
				Relation:      rel,
				SharedNumbers: shared,
				SharedCount:   len(shared),
			})
		}
	}

	return sortedReport(findings)
}

// AnalyseNewSet checks one candidate set against all existing sets.
func AnalyseNewSet(candidate []int, existing [][]int) OverlapReport {
	var findings []OverlapFinding
	candidateSet := toSet(candidate)

	for i, ex := range existing {
		exSet := toSet(ex)
		shared := sharedNumbers(candidateSet, ex)
		smaller := len(candidate)
		if len(ex) < smaller {
			smaller = len(ex)
		}
		if len(shared)*10 < smaller*8 {
			continue
		}
		rel := classify(candidate, ex, candidateSet, exSet, shared)
		findings = append(findings, OverlapFinding{
			SetA:          len(existing), // candidate is logically appended at the end
			SetB:          i,
			Relation:      rel,
			SharedNumbers: shared,
			SharedCount:   len(shared),
		})
	}

	return sortedReport(findings)
}

func sortedReport(findings []OverlapFinding) OverlapReport {
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].SharedCount > findings[j].SharedCount
	})
	return OverlapReport{
		Findings: findings,
		Clean:    len(findings) == 0,
	}
}

func toSet(nums []int) map[int]bool {
	s := make(map[int]bool, len(nums))
	for _, n := range nums {
		s[n] = true
	}
	return s
}

func sharedNumbers(setA map[int]bool, b []int) []int {
	var shared []int
	for _, n := range b {
		if setA[n] {
			shared = append(shared, n)
		}
	}
	return shared
}

func classify(a, b []int, setA, setB map[int]bool, shared []int) OverlapRelation {
	if len(shared) == len(a) && len(shared) == len(b) {
		return RelationEqual
	}
	if len(shared) == len(a) {
		// every number in a is in b → a ⊂ b
		return RelationSubset
	}
	if len(shared) == len(b) {
		// every number in b is in a → a ⊃ b
		return RelationSuperset
	}
	return RelationPartial
}

package domain

import (
	"errors"
	"testing"
)

func mustRounding(t *testing.T, in ScoringStrategy, places int) *Rounding {
	t.Helper()
	r, err := NewRounding(in, places)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func mustPenalty(t *testing.T, in ScoringStrategy, pts float64) *Penalty {
	t.Helper()
	p, err := NewPenalty(in, pts)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestGradeCalculator_KnownSet(t *testing.T) {
	rb := newTestRubric(t)
	// c1: 7/10 (вес 1), c2: 10/20 (вес 3)
	evals := []*Evaluation{
		newTestEval(t, rb, "s1", "c1", 7),
		newTestEval(t, rb, "s1", "c2", 10),
	}

	tests := []struct {
		name     string
		strategy ScoringStrategy
		want     float64
	}{
		{"weighted average", WeightedAverage{}, 55},     // (0.7·1 + 0.5·3)/4
		{"simple sum", SimpleSum{}, 56.666666666666664}, // 17/30
		{"rounded to 0 places", mustRounding(t, WeightedAverage{}, 0), 55},
		{"rounded to 1 place", mustRounding(t, SimpleSum{}, 1), 56.7},
		{"penalty then rounding", mustRounding(t, mustPenalty(t, WeightedAverage{}, 10), 0), 45},
		{"penalty cannot go below zero", mustPenalty(t, WeightedAverage{}, 500), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calc, err := NewGradeCalculator(tc.strategy)
			if err != nil {
				t.Fatal(err)
			}
			got, err := calc.Calculate(rb, evals)
			if err != nil {
				t.Fatal(err)
			}
			if !approx(got.Percent(), tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got.Percent())
			}
		})
	}
}

func TestStrategies_HonourContract(t *testing.T) {
	rb := newTestRubric(t)
	full := []*Evaluation{newTestEval(t, rb, "s1", "c1", 10), newTestEval(t, rb, "s1", "c2", 20)}
	zero := []*Evaluation{newTestEval(t, rb, "s1", "c1", 0), newTestEval(t, rb, "s1", "c2", 0)}

	strategies := map[string]ScoringStrategy{
		"weighted":           WeightedAverage{},
		"sum":                SimpleSum{},
		"rounding(weighted)": mustRounding(t, WeightedAverage{}, 2),
		"penalty(sum)":       mustPenalty(t, SimpleSum{}, 0),
	}
	for name, s := range strategies {
		t.Run(name, func(t *testing.T) {
			hi, err := s.Grade(rb, full)
			if err != nil || !approx(hi.Percent(), 100) {
				t.Fatalf("all max: got %v, %v", hi.Percent(), err)
			}
			lo, err := s.Grade(rb, zero)
			if err != nil || !approx(lo.Percent(), 0) {
				t.Fatalf("all zero: got %v, %v", lo.Percent(), err)
			}
		})
	}
}

func TestGradeCalculator_RejectsInconsistentInput(t *testing.T) {
	rb := newTestRubric(t)
	other := NewRubric()
	c3, _ := NewCriterion("c3", "Чужой критерий", mustMax(t, 10), mustWeight(t, 1))
	if err := other.AddCriterion(c3); err != nil {
		t.Fatal(err)
	}

	calc, err := NewGradeCalculator(WeightedAverage{})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		evals   []*Evaluation
		wantErr error
	}{
		{"no evaluations", nil, ErrNoEvaluations},
		{"nil element", []*Evaluation{nil}, ErrInconsistentEvaluations},
		{"duplicate criterion", []*Evaluation{
			newTestEval(t, rb, "s1", "c1", 5), newTestEval(t, rb, "s1", "c1", 6),
		}, ErrInconsistentEvaluations},
		{"different students", []*Evaluation{
			newTestEval(t, rb, "s1", "c1", 5), newTestEval(t, rb, "s2", "c2", 6),
		}, ErrInconsistentEvaluations},
		{"criterion not in rubric", []*Evaluation{
			newTestEval(t, other, "s1", "c3", 5),
		}, ErrUnknownCriterion},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := calc.Calculate(rb, tc.evals); !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestGradeCalculator_DetectsStaleEvaluation(t *testing.T) {
	// оценка 8/10 была корректной; рубрикатор пересобрали, max стал 5
	old := newTestRubric(t)
	e := newTestEval(t, old, "s1", "c1", 8)

	shrunk := NewRubric()
	c, _ := NewCriterion("c1", "Теория", mustMax(t, 5), mustWeight(t, 1))
	if err := shrunk.AddCriterion(c); err != nil {
		t.Fatal(err)
	}

	calc, _ := NewGradeCalculator(WeightedAverage{})
	if _, err := calc.Calculate(shrunk, []*Evaluation{e}); !errors.Is(err, ErrScoreOutOfRange) {
		t.Fatalf("want ErrScoreOutOfRange, got %v", err)
	}
}

func TestNewGradeCalculator_NilStrategy(t *testing.T) {
	if _, err := NewGradeCalculator(nil); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("got %v", err)
	}
}

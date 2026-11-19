package domain

import (
	"math"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func mustMax(t *testing.T, v float64) MaxScore {
	t.Helper()
	m, err := NewMaxScore(v)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func mustWeight(t *testing.T, v float64) Weight {
	t.Helper()
	w, err := NewWeight(v)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func mustScore(t *testing.T, v float64) Score {
	t.Helper()
	s, err := NewScore(v)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Рубрикатор: c1 (max 10, вес 1) и c2 (max 20, вес 3).
func newTestRubric(t *testing.T) *Rubric {
	t.Helper()
	rb := NewRubric()
	for _, c := range []struct {
		id, name string
		max, w   float64
	}{
		{"c1", "Теория", 10, 1},
		{"c2", "Практика", 20, 3},
	} {
		crit, err := NewCriterion(CriterionID(c.id), c.name, mustMax(t, c.max), mustWeight(t, c.w))
		if err != nil {
			t.Fatal(err)
		}
		if err := rb.AddCriterion(crit); err != nil {
			t.Fatal(err)
		}
	}
	return rb
}

func newTestEval(t *testing.T, rb *Rubric, sid, cid string, score float64) *Evaluation {
	t.Helper()
	f := NewEvaluationFactory(func() time.Time { return fixedNow })
	e, err := f.Create(rb, StudentID(sid), CriterionID(cid), mustScore(t, score), "")
	if err != nil {
		t.Fatal(err)
	}
	return e
}

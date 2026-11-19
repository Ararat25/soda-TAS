package domain

import (
	"errors"
	"testing"
	"time"
)

func TestEvaluationFactory_Create(t *testing.T) {
	rb := newTestRubric(t)
	f := NewEvaluationFactory(func() time.Time { return fixedNow })

	tests := []struct {
		name    string
		student StudentID
		crit    CriterionID
		score   float64
		wantErr error
	}{
		{"middle value", "s1", "c1", 7, nil},
		{"zero", "s1", "c1", 0, nil},
		{"exactly max", "s1", "c1", 10, nil},
		{"12 of 10", "s1", "c1", 12, ErrScoreOutOfRange},
		{"just above max", "s1", "c1", 10.01, ErrScoreOutOfRange},
		{"unknown criterion", "s1", "nope", 5, ErrUnknownCriterion},
		{"empty student id", "", "c1", 5, ErrInvalidValue},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e, err := f.Create(rb, tc.student, tc.crit, mustScore(t, tc.score), "ok")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if tc.wantErr != nil {
				if e != nil {
					t.Fatal("evaluation must not be created on error")
				}
				return
			}
			if !e.At().Equal(fixedNow) || e.Score().Value() != tc.score || e.Comment() != "ok" {
				t.Fatalf("unexpected evaluation: %+v", e)
			}
		})
	}
}

func TestEvaluation_Correct(t *testing.T) {
	rb := newTestRubric(t)
	e := newTestEval(t, rb, "s1", "c1", 7)
	later := fixedNow.Add(time.Hour)

	if err := e.Correct(rb, mustScore(t, 12), "typo", later); !errors.Is(err, ErrScoreOutOfRange) {
		t.Fatalf("want ErrScoreOutOfRange, got %v", err)
	}
	if e.Score().Value() != 7 || !e.At().Equal(fixedNow) {
		t.Fatal("state must not change when correction is rejected")
	}

	if err := e.Correct(rb, mustScore(t, 9), "fixed", later); err != nil {
		t.Fatal(err)
	}
	if e.Score().Value() != 9 || e.Comment() != "fixed" || !e.At().Equal(later) {
		t.Fatalf("correction not applied: %+v", e)
	}
}

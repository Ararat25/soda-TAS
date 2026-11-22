package domain

import (
	"errors"
	"testing"
)

func TestRubric_AddCriterion(t *testing.T) {
	rb := newTestRubric(t)

	tests := []struct {
		name    string
		id      CriterionID
		cname   string
		wantErr error
	}{
		{"new criterion", "c3", "Домашние задания", nil},
		{"duplicate id", "c1", "Другое имя", ErrDuplicateCriterion},
		{"duplicate name, other case", "c9", "ТЕОРИЯ", ErrDuplicateCriterion},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewCriterion(tc.id, tc.cname, mustMax(t, 5), mustWeight(t, 1))
			if err != nil {
				t.Fatal(err)
			}
			if err := rb.AddCriterion(c); !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestNewCriterion_RejectsZeroValueObjects(t *testing.T) {
	if _, err := NewCriterion("c", "name", MaxScore{}, mustWeight(t, 1)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("zero MaxScore: got %v", err)
	}
	if _, err := NewCriterion("c", "name", mustMax(t, 1), Weight{}); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("zero Weight: got %v", err)
	}
	if _, err := NewCriterion("c", "   ", mustMax(t, 1), mustWeight(t, 1)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("blank name: got %v", err)
	}
}

func TestRubric_UnknownCriterion(t *testing.T) {
	rb := newTestRubric(t)
	if _, err := rb.Criterion("nope"); !errors.Is(err, ErrUnknownCriterion) {
		t.Fatalf("got %v", err)
	}
}

func TestRubric_CriteriaReturnsCopy(t *testing.T) {
	rb := newTestRubric(t)
	list := rb.Criteria()
	if len(list) != 2 || list[0].ID() != "c1" || list[1].ID() != "c2" {
		t.Fatalf("unexpected order: %v", list)
	}
	list[0] = Criterion{} // портим копию
	if rb.Criteria()[0].ID() != "c1" {
		t.Fatal("rubric state leaked through Criteria()")
	}
}

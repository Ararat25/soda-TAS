package domain

import (
	"errors"
	"math"
	"testing"
)

func TestValueObjectConstructors(t *testing.T) {
	tests := []struct {
		name    string
		build   func() error
		wantErr bool
	}{
		{"score 0", func() error { _, e := NewScore(0); return e }, false},
		{"score negative", func() error { _, e := NewScore(-1); return e }, true},
		{"score NaN", func() error { _, e := NewScore(math.NaN()); return e }, true},
		{"score +Inf", func() error { _, e := NewScore(math.Inf(1)); return e }, true},
		{"max 0", func() error { _, e := NewMaxScore(0); return e }, true},
		{"max negative", func() error { _, e := NewMaxScore(-5); return e }, true},
		{"max 10", func() error { _, e := NewMaxScore(10); return e }, false},
		{"weight 0", func() error { _, e := NewWeight(0); return e }, true},
		{"weight 2.5", func() error { _, e := NewWeight(2.5); return e }, false},
		{"grade 100", func() error { _, e := NewFinalGrade(100); return e }, false},
		{"grade 100.01", func() error { _, e := NewFinalGrade(100.01); return e }, true},
		{"grade -0.1", func() error { _, e := NewFinalGrade(-0.1); return e }, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.build()
			if tc.wantErr && !errors.Is(err, ErrInvalidValue) {
				t.Fatalf("want ErrInvalidValue, got %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

package domain

import (
	"fmt"
	"math"
)

type (
	StudentID   string // Идентификатор студента
	CriterionID string // Идентификатор критерия
)

// isFinite проверяет, является ли значение float64 конечным числом
func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// MaxScore - максимальный балл критерия (> 0). Value Object.
type MaxScore struct {
	v float64
}

func NewMaxScore(v float64) (MaxScore, error) {
	if !isFinite(v) || v <= 0 {
		return MaxScore{}, fmt.Errorf("%w: max score must be > 0, got %v", ErrInvalidValue, v)
	}
	return MaxScore{v: v}, nil
}

func (m MaxScore) Value() float64 { return m.v }

// Weight - вклад критерия в итог (> 0). Value Object.
type Weight struct{ v float64 }

func NewWeight(v float64) (Weight, error) {
	if !isFinite(v) || v <= 0 {
		return Weight{}, fmt.Errorf("%w: weight must be > 0, got %v", ErrInvalidValue, v)
	}
	return Weight{v: v}, nil
}

func (w Weight) Value() float64 { return w.v }

// Score - выставленный балл (>= 0). Value Object.
type Score struct{ v float64 }

func NewScore(v float64) (Score, error) {
	if !isFinite(v) || v < 0 {
		return Score{}, fmt.Errorf("%w: score must be >= 0, got %v", ErrInvalidValue, v)
	}
	return Score{v: v}, nil
}

func (s Score) Value() float64 { return s.v }

// FinalGrade - итог по студенту в процентах [0, 100]. Value Object.
type FinalGrade struct{ percent float64 }

func NewFinalGrade(percent float64) (FinalGrade, error) {
	if !isFinite(percent) || percent < 0 || percent > 100 {
		return FinalGrade{}, fmt.Errorf("%w: grade must be in [0, 100], got %v", ErrInvalidValue, percent)
	}
	return FinalGrade{percent: percent}, nil
}

func (g FinalGrade) Percent() float64 { return g.percent }

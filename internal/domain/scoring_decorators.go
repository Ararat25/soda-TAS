package domain

import (
	"fmt"
	"math"
)

// Rounding — [GoF Decorator]: сам является ScoringStrategy и оборачивает другую.
type Rounding struct {
	inner  ScoringStrategy
	places int
}

// NewRounding создаёт декоратор округления итоговой оценки
func NewRounding(inner ScoringStrategy, places int) (*Rounding, error) {
	if inner == nil || places < 0 || places > 6 {
		return nil, fmt.Errorf("%w: rounding needs a strategy and 0..6 places", ErrInvalidValue)
	}
	return &Rounding{inner: inner, places: places}, nil
}

// Grade округляет итог, полученный от вложенной стратегии
func (r *Rounding) Grade(rb *Rubric, evals []*Evaluation) (FinalGrade, error) {
	g, err := r.inner.Grade(rb, evals)
	if err != nil {
		return FinalGrade{}, err
	}

	k := math.Pow(10, float64(r.places))

	return NewFinalGrade(clampPercent(math.Round(g.percent*k) / k))
}

// Penalty — [GoF Decorator]: вычитает фиксированное число процентных пунктов (не ниже 0).
type Penalty struct {
	inner  ScoringStrategy
	points float64
}

// NewPenalty создаёт декоратор штрафа
func NewPenalty(inner ScoringStrategy, points float64) (*Penalty, error) {
	if inner == nil || !isFinite(points) || points < 0 {
		return nil, fmt.Errorf("%w: penalty needs a strategy and points >= 0", ErrInvalidValue)
	}
	return &Penalty{inner: inner, points: points}, nil
}

// Grade вычитает штраф из итога вложенной стратегии
func (p *Penalty) Grade(rb *Rubric, evals []*Evaluation) (FinalGrade, error) {
	g, err := p.inner.Grade(rb, evals)

	if err != nil {
		return FinalGrade{}, err
	}

	return NewFinalGrade(clampPercent(g.percent - p.points))
}

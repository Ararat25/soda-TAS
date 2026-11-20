package domain

import "math"

// ScoringStrategy — [GoF Strategy]: сменяемое правило расчёта итога
type ScoringStrategy interface {
	Grade(rb *Rubric, evals []*Evaluation) (FinalGrade, error)
}

// clamp защищает от накопленной погрешности float (100.00000000000001)
func clampPercent(p float64) float64 { return math.Max(0, math.Min(100, p)) }

// WeightedAverage суммирует отношения балла студента по критерию к максимуму, умноженные
// на веса критериев, делит на сумму весов и умножает на 100.
type WeightedAverage struct{}

// Grade считает итог как взвешенное среднее по критериям
func (WeightedAverage) Grade(rb *Rubric, evals []*Evaluation) (FinalGrade, error) {
	var sum, wsum float64

	for _, e := range evals {
		c, err := rb.Criterion(e.criterionID)
		if err != nil {
			return FinalGrade{}, err
		}

		sum += e.score.v / c.max.v * c.weight.v
		wsum += c.weight.v
	}

	if wsum == 0 {
		return FinalGrade{}, ErrNoEvaluations
	}

	return NewFinalGrade(clampPercent(sum / wsum * 100))
}

// SimpleSum суммирует баллы студента по всем критериям, делит на сумму максимумов и умножает на 100
type SimpleSum struct{}

// Grade считает итог как сумму баллов, делённую на сумму максимумов
func (SimpleSum) Grade(rb *Rubric, evals []*Evaluation) (FinalGrade, error) {
	var got, total float64

	for _, e := range evals {
		c, err := rb.Criterion(e.criterionID)
		if err != nil {
			return FinalGrade{}, err
		}

		got += e.score.v
		total += c.max.v
	}

	if total == 0 {
		return FinalGrade{}, ErrNoEvaluations
	}

	return NewFinalGrade(clampPercent(got / total * 100))
}

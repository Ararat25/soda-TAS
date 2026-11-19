package domain

import "fmt"

// GradeCalculator — Domain Service: проверяет согласованность набора
// оценок с текущим рубрикатором и делегирует расчет стратегии.
type GradeCalculator struct{ strategy ScoringStrategy }

// NewGradeCalculator конструктор GradeCalculator
func NewGradeCalculator(s ScoringStrategy) (*GradeCalculator, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: scoring strategy is nil", ErrInvalidValue)
	}
	return &GradeCalculator{strategy: s}, nil
}

// Calculate считает оценку используя выбранную стратению
func (c *GradeCalculator) Calculate(rb *Rubric, evals []*Evaluation) (FinalGrade, error) {
	if len(evals) == 0 {
		return FinalGrade{}, ErrNoEvaluations
	}

	if evals[0] == nil {
		return FinalGrade{}, fmt.Errorf("%w: nil evaluation", ErrInconsistentEvaluations)
	}

	owner := evals[0].studentID
	seen := make(map[CriterionID]struct{}, len(evals))

	for _, e := range evals {
		if e == nil {
			return FinalGrade{}, fmt.Errorf("%w: nil evaluation", ErrInconsistentEvaluations)
		}

		if e.studentID != owner {
			return FinalGrade{}, fmt.Errorf("%w: evaluations of different students", ErrInconsistentEvaluations)
		}

		if _, dup := seen[e.criterionID]; dup {
			return FinalGrade{}, fmt.Errorf("%w: two evaluations for criterion %q", ErrInconsistentEvaluations, e.criterionID)
		}

		seen[e.criterionID] = struct{}{}

		crit, err := rb.Criterion(e.criterionID)
		if err != nil {
			return FinalGrade{}, err
		}

		// рубрикатор мог измениться после выставления оценки
		if err := crit.Validate(e.score); err != nil {
			return FinalGrade{}, err
		}
	}

	return c.strategy.Grade(rb, evals)
}

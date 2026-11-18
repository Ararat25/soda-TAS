package domain

import (
	"fmt"
	"time"
)

// EvaluationFactory — [GoF Factory]: единая точка создания оценки с инвариантами.
// Время необходимо, чтобы тесты были детерминированными.
type EvaluationFactory struct {
	now func() time.Time
}

func NewEvaluationFactory(now func() time.Time) *EvaluationFactory {
	if now == nil {
		now = time.Now
	}

	return &EvaluationFactory{now: now}
}

func (f *EvaluationFactory) Create(rb *Rubric, sid StudentID, cid CriterionID, s Score, comment string) (*Evaluation, error) {
	if sid == "" {
		return nil, fmt.Errorf("%w: student id is empty", ErrInvalidValue)
	}

	crit, err := rb.Criterion(cid) // оценить можно только критерий из рубрикатора
	if err != nil {
		return nil, err
	}

	if err := crit.Validate(s); err != nil { // балл в пределах критерия
		return nil, err
	}

	return &Evaluation{studentID: sid, criterionID: cid, score: s, comment: comment, at: f.now()}, nil
}

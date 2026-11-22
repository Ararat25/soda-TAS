package domain

import "time"

// Evaluation — оценивание (этому студенту по этому критерию поставили такой балл)
// Создаётся только через EvaluationFactory.
type Evaluation struct {
	studentID   StudentID
	criterionID CriterionID
	score       Score
	comment     string
	at          time.Time
}

// StudentID возвращает идентификатор студента
func (e *Evaluation) StudentID() StudentID { return e.studentID }

// CriterionID возвращает идентификатор критерия
func (e *Evaluation) CriterionID() CriterionID { return e.criterionID }

// Score возвращает выставленный балл
func (e *Evaluation) Score() Score { return e.score }

// Comment возвращает комментарий к оценке
func (e *Evaluation) Comment() string { return e.comment }

// At возвращает время выставления оценки
func (e *Evaluation) At() time.Time { return e.at }

// Correct исправляет оценку и проверяет инварианты. Состояние меняется только после успешной проверки.
func (e *Evaluation) Correct(rb *Rubric, s Score, comment string, at time.Time) error {
	crit, err := rb.Criterion(e.criterionID)
	if err != nil {
		return err
	}

	if err := crit.Validate(s); err != nil {
		return err
	}

	e.score, e.comment, e.at = s, comment, at

	return nil
}

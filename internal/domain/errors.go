package domain

import "errors"

// Доменные ошибки. Оборачиваются через %w, проверяются через errors.Is.
var (
	ErrInvalidValue            = errors.New("invalid value")
	ErrScoreOutOfRange         = errors.New("score is outside [0, max_score]")
	ErrUnknownCriterion        = errors.New("criterion is not in the rubric")
	ErrDuplicateCriterion      = errors.New("criterion already exists")
	ErrNotFound                = errors.New("not found")
	ErrNoEvaluations           = errors.New("no evaluations to calculate a grade from")
	ErrInconsistentEvaluations = errors.New("evaluations are inconsistent")
)

package repository

import "github.com/Ararat25/soda-TAS/internal/domain"

type StudentRepository interface {
	Save(*domain.Student) error
	ByID(domain.StudentID) (*domain.Student, error)
	List() ([]*domain.Student, error)
}

type RubricRepository interface {
	Get() (*domain.Rubric, error)
	Save(*domain.Rubric) error
}

type EvaluationRepository interface {
	Save(*domain.Evaluation) error
	Find(domain.StudentID, domain.CriterionID) (*domain.Evaluation, error)
	ByStudent(domain.StudentID) ([]*domain.Evaluation, error)
	Delete(domain.StudentID, domain.CriterionID) error
}

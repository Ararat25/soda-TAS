package usecase

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Ararat25/soda-TAS/internal/domain"
	"github.com/Ararat25/soda-TAS/internal/repository"
)

const (
	studentPrefix   = "s" // префикс id студентов для репозитория
	criterionPrefix = "c" // префикс id критериев для репозитория
)

// IDGenerator — порт: откуда берутся идентификаторы
type IDGenerator interface {
	NextID(prefix string) string
}

// Clock — порт времени (подменяется в тестах).
type Clock func() time.Time

// Service — единая точка сценариев. Service является [GoF Facade] для консоли.
type Service struct {
	students repository.StudentRepository
	rubrics  repository.RubricRepository
	evals    repository.EvaluationRepository
	factory  *domain.EvaluationFactory
	calc     *domain.GradeCalculator
	ids      IDGenerator
	now      Clock
}

// NewService создаёт сервис сценариев
func NewService(
	students repository.StudentRepository,
	rubrics repository.RubricRepository,
	evals repository.EvaluationRepository,
	factory *domain.EvaluationFactory,
	calc *domain.GradeCalculator,
	ids IDGenerator,
	now Clock,
) *Service {
	return &Service{
		students: students,
		rubrics:  rubrics,
		evals:    evals,
		factory:  factory,
		calc:     calc,
		ids:      ids,
		now:      now}
}

// AddStudent добавляет нового студента в репозиторий
func (s *Service) AddStudent(name, group string) (domain.StudentID, error) {
	id := domain.StudentID(s.ids.NextID(studentPrefix))

	st, err := domain.NewStudent(id, name, group)
	if err != nil {
		return "", err
	}

	err = s.students.Save(st)
	if err != nil {
		return "", err
	}

	return id, nil
}

// ListStudents возвращает копию списока студентов
func (s *Service) ListStudents() ([]*domain.Student, error) {
	return s.students.List()
}

// AddCriterion добавляет новый критерий в рубрикатор
func (s *Service) AddCriterion(name string, maxScore, weight float64) (domain.CriterionID, error) {
	m, err := domain.NewMaxScore(maxScore)
	if err != nil {
		return "", err
	}

	w, err := domain.NewWeight(weight)
	if err != nil {
		return "", err
	}

	id := domain.CriterionID(s.ids.NextID(criterionPrefix))
	crit, err := domain.NewCriterion(id, name, m, w)
	if err != nil {
		return "", err
	}

	rb, err := s.rubrics.Get()
	if err != nil {
		return "", err
	}

	err = rb.AddCriterion(crit)
	if err != nil {
		return "", err
	}

	err = s.rubrics.Save(rb)
	if err != nil {
		return "", err
	}
	return id, nil
}

// ListCriteria возвращает копии текущих критериев
func (s *Service) ListCriteria() ([]domain.Criterion, error) {
	rb, err := s.rubrics.Get()
	if err != nil {
		return nil, err
	}

	return rb.Criteria(), nil
}

// SetScore выставляет оценку. Инварианты диапазона и рубрикатора проверяет фабрика.
func (s *Service) SetScore(sid domain.StudentID, cid domain.CriterionID, score float64, comment string) error {
	_, err := s.students.ByID(sid)
	if err != nil {
		return err
	}

	sc, err := domain.NewScore(score)
	if err != nil {
		return err
	}

	rb, err := s.rubrics.Get()
	if err != nil {
		return err
	}

	ev, err := s.factory.Create(rb, sid, cid, sc, comment)
	if err != nil {
		return err
	}

	_, err = s.evals.Find(sid, cid)
	switch {
	case err == nil:
		return fmt.Errorf("%w: student %q, criterion %q (use correct)", domain.ErrAlreadyEvaluated, sid, cid)
	case !errors.Is(err, domain.ErrNotFound):
		return err
	}

	return s.evals.Save(ev)
}

// CorrectScore изменяет оценку студента по критерию
func (s *Service) CorrectScore(sid domain.StudentID, cid domain.CriterionID, score float64, comment string) error {
	sc, err := domain.NewScore(score)
	if err != nil {
		return err
	}

	rb, err := s.rubrics.Get()
	if err != nil {
		return err
	}

	ev, err := s.evals.Find(sid, cid)
	if err != nil {
		return err
	}

	err = ev.Correct(rb, sc, comment, s.now())
	if err != nil {
		return err
	}

	return s.evals.Save(ev)
}

// CancelScore удаляет оценку студента по критерию
func (s *Service) CancelScore(sid domain.StudentID, cid domain.CriterionID) error {
	return s.evals.Delete(sid, cid)
}

// ListEvaluations возвращает список оценок студента
func (s *Service) ListEvaluations(sid domain.StudentID) ([]*domain.Evaluation, error) {
	_, err := s.students.ByID(sid)
	if err != nil {
		return nil, err
	}

	return s.evals.ByStudent(sid)
}

// FinalGrade считает итоговую оценку студента
func (s *Service) FinalGrade(sid domain.StudentID) (domain.FinalGrade, error) {
	_, err := s.students.ByID(sid)
	if err != nil {
		return domain.FinalGrade{}, err
	}

	rb, err := s.rubrics.Get()
	if err != nil {
		return domain.FinalGrade{}, err
	}

	evals, err := s.evals.ByStudent(sid)
	if err != nil {
		return domain.FinalGrade{}, err
	}

	return s.calc.Calculate(rb, evals)
}

// Report собирает итоговые оценки по всем студентам
func (s *Service) Report() ([]ReportRow, error) {
	students, err := s.students.List()
	if err != nil {
		return nil, err
	}

	rows := make([]ReportRow, 0, len(students))
	for _, st := range students {
		row := ReportRow{StudentID: st.ID(), StudentName: st.Name(), Group: st.Group()}

		g, err := s.FinalGrade(st.ID())
		switch {
		case err == nil:
			row.Graded, row.Grade = true, g
		case !errors.Is(err, domain.ErrNoEvaluations):
			return nil, err
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// ExportFinal экспортирует строки из Report через Exporter
func (s *Service) ExportFinal(w io.Writer, e Exporter) error {
	rows, err := s.Report()
	if err != nil {
		return err
	}

	return e.Export(w, rows)
}

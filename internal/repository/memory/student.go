package memory

import (
	"fmt"
	"sync"

	"github.com/Ararat25/soda-TAS/internal/domain"
	"github.com/Ararat25/soda-TAS/internal/repository"
)

var _ repository.StudentRepository = (*StudentRepo)(nil)

// StudentRepo — in-memory реализация StudentRepository
type StudentRepo struct {
	mu    sync.RWMutex
	byID  map[domain.StudentID]*domain.Student
	order []domain.StudentID
}

// NewStudentRepo создаёт in-memory репозиторий студентов
func NewStudentRepo() *StudentRepo {
	return &StudentRepo{byID: make(map[domain.StudentID]*domain.Student)}
}

// Save сохраняет студента
func (r *StudentRepo) Save(s *domain.Student) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[s.ID()]; !ok {
		r.order = append(r.order, s.ID())
	}

	r.byID[s.ID()] = s

	return nil
}

// ByID возвращает студента по идентификатору
func (r *StudentRepo) ByID(id domain.StudentID) (*domain.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("%w: student %q", domain.ErrNotFound, id)
	}

	return s, nil
}

// List возвращает студентов в порядке добавления
func (r *StudentRepo) List() ([]*domain.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.Student, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.byID[id])
	}

	return out, nil
}

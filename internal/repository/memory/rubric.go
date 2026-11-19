package memory

import (
	"sync"

	"github.com/Ararat25/soda-TAS/internal/domain"
	"github.com/Ararat25/soda-TAS/internal/repository"
)

var _ repository.RubricRepository = (*RubricRepo)(nil)

// RubricRepo хранит единственный рубрикатор приложения
type RubricRepo struct {
	mu sync.RWMutex
	rb *domain.Rubric
}

func NewRubricRepo() *RubricRepo { return &RubricRepo{rb: domain.NewRubric()} }

// Get возвращает текущий рубрикатор
func (r *RubricRepo) Get() (*domain.Rubric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.rb, nil
}

// Save сохраняет новый рубрикатор
func (r *RubricRepo) Save(rb *domain.Rubric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rb = rb

	return nil
}

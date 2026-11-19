package memory

import (
	"fmt"
	"sync"

	"github.com/Ararat25/soda-TAS/internal/domain"
	"github.com/Ararat25/soda-TAS/internal/repository"
)

var _ repository.EvaluationRepository = (*EvaluationRepo)(nil)

type evalKey struct {
	student   domain.StudentID
	criterion domain.CriterionID
}

type EvaluationRepo struct {
	mu    sync.RWMutex
	items map[evalKey]*domain.Evaluation
	order []evalKey
}

func NewEvaluationRepo() *EvaluationRepo {
	return &EvaluationRepo{items: make(map[evalKey]*domain.Evaluation)}
}

func (r *EvaluationRepo) Save(e *domain.Evaluation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	k := evalKey{e.StudentID(), e.CriterionID()}
	if _, ok := r.items[k]; !ok {
		r.order = append(r.order, k)
	}

	r.items[k] = e

	return nil
}

func (r *EvaluationRepo) Find(sid domain.StudentID, cid domain.CriterionID) (*domain.Evaluation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.items[evalKey{sid, cid}]
	if !ok {
		return nil, fmt.Errorf("%w: evaluation of student %q by criterion %q", domain.ErrNotFound, sid, cid)
	}

	return e, nil
}

func (r *EvaluationRepo) ByStudent(sid domain.StudentID) ([]*domain.Evaluation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []*domain.Evaluation
	for _, k := range r.order {
		if k.student == sid {
			out = append(out, r.items[k])
		}
	}

	return out, nil
}

func (r *EvaluationRepo) Delete(sid domain.StudentID, cid domain.CriterionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	k := evalKey{sid, cid}
	if _, ok := r.items[k]; !ok {
		return fmt.Errorf("%w: evaluation of student %q by criterion %q", domain.ErrNotFound, sid, cid)
	}

	delete(r.items, k)

	for i, ok := range r.order {
		if ok == k {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}

	return nil
}

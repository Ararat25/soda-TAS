package domain

import (
	"fmt"
	"strings"
)

// Criterion — критерий, за который ставится балл
type Criterion struct {
	id     CriterionID
	name   string
	max    MaxScore
	weight Weight
}

// NewCriterion - конструктор Criterion
func NewCriterion(id CriterionID, name string, m MaxScore, w Weight) (Criterion, error) {
	name = strings.TrimSpace(name)

	switch {
	case id == "":
		return Criterion{}, fmt.Errorf("%w: criterion id is empty", ErrInvalidValue)
	case name == "":
		return Criterion{}, fmt.Errorf("%w: criterion name is empty", ErrInvalidValue)
	case m.v <= 0:
		return Criterion{}, fmt.Errorf("%w: criterion max score is not set", ErrInvalidValue)
	case w.v <= 0:
		return Criterion{}, fmt.Errorf("%w: criterion weight is not set", ErrInvalidValue)
	}

	return Criterion{id: id, name: name, max: m, weight: w}, nil
}

func (c Criterion) ID() CriterionID { return c.id }
func (c Criterion) Name() string    { return c.name }
func (c Criterion) Max() MaxScore   { return c.max }
func (c Criterion) Weight() Weight  { return c.weight }

// Validate — балл не выходит за пределы критерия
func (c Criterion) Validate(s Score) error {
	if s.v > c.max.v {
		return fmt.Errorf("%w: got %v, max %v (criterion %q)", ErrScoreOutOfRange, s.v, c.max.v, c.name)
	}

	return nil
}

// Rubric — рубрикатор, критерии существуют только внутри него
type Rubric struct {
	criteria map[CriterionID]Criterion
	order    []CriterionID
}

// NewRubric - конструктор Rubric
func NewRubric() *Rubric {
	return &Rubric{criteria: make(map[CriterionID]Criterion)}
}

// AddCriterion - добавляет новый критерий в рубрикатор
func (r *Rubric) AddCriterion(c Criterion) error {
	if _, ok := r.criteria[c.id]; ok {
		return fmt.Errorf("%w: id %q", ErrDuplicateCriterion, c.id)
	}

	for _, ex := range r.criteria {
		if strings.EqualFold(ex.name, c.name) {
			return fmt.Errorf("%w: name %q", ErrDuplicateCriterion, c.name)
		}
	}

	r.criteria[c.id] = c
	r.order = append(r.order, c.id)

	return nil
}

// Criterion - возвращает критерий по CriterionID
func (r *Rubric) Criterion(id CriterionID) (Criterion, error) {
	c, ok := r.criteria[id]
	if !ok {
		return Criterion{}, fmt.Errorf("%w: %q", ErrUnknownCriterion, id)
	}

	return c, nil
}

// Criteria возвращает копию критериев в порядке добавления
func (r *Rubric) Criteria() []Criterion {
	out := make([]Criterion, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.criteria[id])
	}

	return out
}

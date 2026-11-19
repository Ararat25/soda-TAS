package domain

import (
	"fmt"
	"strings"
)

// Student — Entity: идентичность по ID, ФИО и группа могут меняться.
type Student struct {
	id    StudentID
	name  string
	group string
}

func NewStudent(id StudentID, name, group string) (*Student, error) {
	name = strings.TrimSpace(name)
	if id == "" {
		return nil, fmt.Errorf("%w: student id is empty", ErrInvalidValue)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: student name is empty", ErrInvalidValue)
	}
	return &Student{id: id, name: name, group: strings.TrimSpace(group)}, nil
}

func (s *Student) ID() StudentID { return s.id }
func (s *Student) Name() string  { return s.name }
func (s *Student) Group() string { return s.group }

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

// NewStudent создаёт нового студента
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

// ID возвращает идентификатор студента
func (s *Student) ID() StudentID { return s.id }

// Name возвращает ФИО студента
func (s *Student) Name() string { return s.name }

// Group возвращает группу студента
func (s *Student) Group() string { return s.group }

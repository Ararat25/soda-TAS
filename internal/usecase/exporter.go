package usecase

import (
	"io"

	"github.com/Ararat25/soda-TAS/internal/domain"
)

// ReportRow — строка выгрузки итоговых оценок студентов
type ReportRow struct {
	StudentID   domain.StudentID
	StudentName string
	Group       string
	Graded      bool // Graded=false - у студента ещё нет оценок
	Grade       domain.FinalGrade
}

// Exporter — порт: выгрузчик данных в разнвх форматах
type Exporter interface {
	Export(w io.Writer, rows []ReportRow) error
}

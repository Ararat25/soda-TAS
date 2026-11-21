package export

import (
	"encoding/json"
	"io"

	"github.com/Ararat25/soda-TAS/internal/usecase"
)

var _ usecase.Exporter = (*JSON)(nil)

// JSON — экспорт итогов в JSON
type JSON struct{}

// NewJSON создаёт экспортёр JSON
func NewJSON() *JSON { return &JSON{} }

// jsonRow — строка итогов для JSON-выгрузки
type jsonRow struct {
	StudentID    string   `json:"student_id"`
	Name         string   `json:"name"`
	Group        string   `json:"group"`
	Graded       bool     `json:"graded"`
	GradePercent *float64 `json:"grade_percent"` // null, если оценок ещё нет
}

// Export выгружает строки итогов в JSON
func (*JSON) Export(w io.Writer, rows []usecase.ReportRow) error {
	out := make([]jsonRow, 0, len(rows))
	for _, r := range rows {
		jr := jsonRow{
			StudentID: string(r.StudentID),
			Name:      r.StudentName,
			Group:     r.Group,
			Graded:    r.Graded}

		if r.Graded {
			p := r.Grade.Percent()
			jr.GradePercent = &p
		}

		out = append(out, jr)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	return enc.Encode(out)
}

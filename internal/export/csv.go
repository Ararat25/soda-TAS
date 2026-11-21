package export

import (
	"encoding/csv"
	"io"
	"strconv"

	"github.com/Ararat25/soda-TAS/internal/usecase"
)

var _ usecase.Exporter = (*CSV)(nil)

// CSV — экспорт итогов в CSV
type CSV struct{}

// NewCSV создаёт экспортёр CSV
func NewCSV() *CSV { return &CSV{} }

// Export выгружает строки итогов в CSV
func (*CSV) Export(w io.Writer, rows []usecase.ReportRow) error {
	_, err := w.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	err = cw.Write([]string{"student_id", "name", "group", "grade_percent"})
	if err != nil {
		return err
	}

	for _, r := range rows {
		grade := "" // у студента без оценок итога нет
		if r.Graded {
			grade = strconv.FormatFloat(r.Grade.Percent(), 'f', 2, 64)
		}

		err = cw.Write([]string{string(r.StudentID), r.StudentName, r.Group, grade})
		if err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

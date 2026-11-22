package usecase_test

import (
	"errors"
	"io"
	"math"
	"testing"
	"time"

	"github.com/Ararat25/soda-TAS/internal/domain"
	"github.com/Ararat25/soda-TAS/internal/idgen"
	"github.com/Ararat25/soda-TAS/internal/repository/memory"
	"github.com/Ararat25/soda-TAS/internal/usecase"
)

var fixed = time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)

type env struct {
	svc    *usecase.Service
	sid    domain.StudentID
	c1, c2 domain.CriterionID // c1: max 10, вес 1; c2: max 20, вес 3
}

func newEnv(t *testing.T) env {
	t.Helper()
	calc, err := domain.NewGradeCalculator(domain.WeightedAverage{})
	if err != nil {
		t.Fatal(err)
	}
	clock := func() time.Time { return fixed }
	svc := usecase.NewService(
		memory.NewStudentRepo(), memory.NewRubricRepo(), memory.NewEvaluationRepo(),
		domain.NewEvaluationFactory(clock), calc, idgen.New(), usecase.Clock(clock),
	)
	var e env
	e.svc = svc
	if e.sid, err = svc.AddStudent("Иван Петров", "Б01-001"); err != nil {
		t.Fatal(err)
	}
	if e.c1, err = svc.AddCriterion("Теория", 10, 1); err != nil {
		t.Fatal(err)
	}
	if e.c2, err = svc.AddCriterion("Практика", 20, 3); err != nil {
		t.Fatal(err)
	}
	return e
}

func TestSetScore_Invariants(t *testing.T) {
	tests := []struct {
		name    string
		student domain.StudentID
		crit    func(env) domain.CriterionID
		score   float64
		wantErr error
	}{
		{"valid", "s1", func(e env) domain.CriterionID { return e.c1 }, 7, nil},
		{"12 of 10", "s1", func(e env) domain.CriterionID { return e.c1 }, 12, domain.ErrScoreOutOfRange},
		{"negative score", "s1", func(e env) domain.CriterionID { return e.c1 }, -1, domain.ErrInvalidValue},
		{"criterion not in rubric", "s1", func(env) domain.CriterionID { return "c99" }, 5, domain.ErrUnknownCriterion},
		{"unknown student", "s99", func(e env) domain.CriterionID { return e.c1 }, 5, domain.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t)
			err := e.svc.SetScore(tc.student, tc.crit(e), tc.score, "")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			evs, _ := e.svc.ListEvaluations(e.sid)
			wantCount := 0
			if tc.wantErr == nil {
				wantCount = 1
			}
			if len(evs) != wantCount {
				t.Fatalf("stored evaluations: want %d, got %d", wantCount, len(evs))
			}
		})
	}
}

func TestEvaluationLifecycle(t *testing.T) {
	e := newEnv(t)

	if err := e.svc.SetScore(e.sid, e.c1, 7, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.SetScore(e.sid, e.c1, 8, ""); !errors.Is(err, domain.ErrAlreadyEvaluated) {
		t.Fatalf("second set: want ErrAlreadyEvaluated, got %v", err)
	}
	if err := e.svc.CorrectScore(e.sid, e.c1, 12, ""); !errors.Is(err, domain.ErrScoreOutOfRange) {
		t.Fatalf("correct out of range: got %v", err)
	}
	if err := e.svc.CorrectScore(e.sid, e.c1, 9, "fixed"); err != nil {
		t.Fatal(err)
	}
	evs, _ := e.svc.ListEvaluations(e.sid)
	if len(evs) != 1 || evs[0].Score().Value() != 9 || evs[0].Comment() != "fixed" {
		t.Fatalf("unexpected evaluations: %+v", evs)
	}

	if err := e.svc.CancelScore(e.sid, e.c1); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.CancelScore(e.sid, e.c1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("second cancel: got %v", err)
	}
	if err := e.svc.CorrectScore(e.sid, e.c1, 5, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("correct after cancel: got %v", err)
	}
}

func TestFinalGrade(t *testing.T) {
	e := newEnv(t)

	if _, err := e.svc.FinalGrade(e.sid); !errors.Is(err, domain.ErrNoEvaluations) {
		t.Fatalf("no evaluations: got %v", err)
	}
	if _, err := e.svc.FinalGrade("s99"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown student: got %v", err)
	}

	mustSet(t, e.svc, e.sid, e.c1, 7)  // 0.7 · вес 1
	mustSet(t, e.svc, e.sid, e.c2, 10) // 0.5 · вес 3
	g, err := e.svc.FinalGrade(e.sid)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(g.Percent()-55) > 1e-9 { // (0.7 + 1.5) / 4
		t.Fatalf("want 55, got %v", g.Percent())
	}
}

type captureExporter struct{ rows []usecase.ReportRow }

func (c *captureExporter) Export(_ io.Writer, rows []usecase.ReportRow) error {
	c.rows = rows
	return nil
}

func TestExportFinal_IncludesUngradedStudents(t *testing.T) {
	e := newEnv(t)
	if _, err := e.svc.AddStudent("Анна Смирнова", "Б01-002"); err != nil {
		t.Fatal(err)
	}
	mustSet(t, e.svc, e.sid, e.c1, 7)
	mustSet(t, e.svc, e.sid, e.c2, 10)

	exp := &captureExporter{}
	if err := e.svc.ExportFinal(io.Discard, exp); err != nil {
		t.Fatal(err)
	}
	if len(exp.rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(exp.rows))
	}
	if !exp.rows[0].Graded || math.Abs(exp.rows[0].Grade.Percent()-55) > 1e-9 {
		t.Fatalf("first row: %+v", exp.rows[0])
	}
	if exp.rows[1].Graded {
		t.Fatalf("second student has no evaluations, row must be ungraded: %+v", exp.rows[1])
	}
}

func TestAddCriterion_RejectsInvalidInput(t *testing.T) {
	e := newEnv(t)
	tests := []struct {
		name    string
		cname   string
		max, w  float64
		wantErr error
	}{
		{"duplicate name", "теория", 5, 1, domain.ErrDuplicateCriterion},
		{"zero max", "Новый", 0, 1, domain.ErrInvalidValue},
		{"zero weight", "Новый", 5, 0, domain.ErrInvalidValue},
		{"blank name", "  ", 5, 1, domain.ErrInvalidValue},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := e.svc.AddCriterion(tc.cname, tc.max, tc.w); !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
		})
	}
	if list, _ := e.svc.ListCriteria(); len(list) != 2 {
		t.Fatalf("rubric must stay unchanged, got %d criteria", len(list))
	}
}

func mustSet(t *testing.T, svc *usecase.Service, sid domain.StudentID, cid domain.CriterionID, score float64) {
	t.Helper()
	if err := svc.SetScore(sid, cid, score, ""); err != nil {
		t.Fatal(err)
	}
}

package cli

import (
	"errors"
	"fmt"

	"github.com/Ararat25/soda-TAS/internal/domain"
)

// Demo — сквозной сценарий
func (c *CLI) Demo() error {
	step := func(title string) error {
		_, err := fmt.Fprintf(c.out, "\n=== %s ===\n", title)
		return err
	}

	err := step("1. Студенты")
	if err != nil {
		return err
	}

	ivan, err := c.demoStudent("Иван Петров", "Б01-001")
	if err != nil {
		return err
	}

	anna, err := c.demoStudent("Анна Смирнова", "Б01-002")
	if err != nil {
		return err
	}

	_, err = c.demoStudent("Пётр Сидоров", "Б01-003")
	if err != nil {
		return err
	}

	err = step("2. Рубрикатор")
	if err != nil {
		return err
	}

	theory, err := c.demoCriterion("Теория", 10, 1)
	if err != nil {
		return err
	}

	practice, err := c.demoCriterion("Практика", 20, 3)
	if err != nil {
		return err
	}

	project, err := c.demoCriterion("Проект", 30, 2)
	if err != nil {
		return err
	}

	err = step("3. Выставление оценок")
	if err != nil {
		return err
	}

	scores := []struct {
		sid   domain.StudentID
		cid   domain.CriterionID
		score float64
	}{
		{ivan, theory, 7}, {ivan, practice, 15}, {ivan, project, 24},
		{anna, theory, 9}, {anna, practice, 18},
	}

	for _, s := range scores {
		err = c.svc.SetScore(s.sid, s.cid, s.score, "")
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(c.out, "  + %s / %s = %g\n", s.sid, s.cid, s.score)
		if err != nil {
			return err
		}
	}

	err = step("4. Модель не даёт нарушить правила домена")
	if err != nil {
		return err
	}

	err = c.expectRejected("12 баллов из 10", domain.ErrScoreOutOfRange,
		c.svc.SetScore(anna, theory, 12, ""))
	if err != nil {
		return err
	}

	err = c.expectRejected("критерий не из рубрикатора", domain.ErrUnknownCriterion,
		c.svc.SetScore(anna, "c99", 5, ""))
	if err != nil {
		return err
	}

	err = c.expectRejected("повторная оценка той же пары", domain.ErrAlreadyEvaluated,
		c.svc.SetScore(anna, theory, 5, ""))
	if err != nil {
		return err
	}

	err = step("5. Исправление и отмена")
	if err != nil {
		return err
	}

	err = c.svc.CorrectScore(ivan, theory, 8, "пересмотрено")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.out, "  исправлено: %s / %s → 8\n", ivan, theory)
	if err != nil {
		return err
	}

	err = c.svc.CancelScore(anna, practice)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.out, "  отменено:   %s / %s\n", anna, practice)
	if err != nil {
		return err
	}

	err = c.svc.SetScore(anna, practice, 17, "после пересдачи")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.out, "  выставлено заново: %s / %s = 17\n", anna, practice)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.out, "Оценки Ивана:\n")
	if err != nil {
		return err
	}

	err = c.printEvaluations(ivan)
	if err != nil {
		return err
	}

	err = step("6. Итоги")
	if err != nil {
		return err
	}

	students, err := c.svc.ListStudents()
	if err != nil {
		return err
	}

	for _, s := range students {
		g, err := c.svc.FinalGrade(s.ID())
		switch {
		case err == nil:
			_, wErr := fmt.Fprintf(c.out, "  %-18s %.1f%%\n", s.Name(), g.Percent())
			if wErr != nil {
				return wErr
			}
		case errors.Is(err, domain.ErrNoEvaluations):
			_, wErr := fmt.Fprintf(c.out, "  %-18s нет оценок\n", s.Name())
			if wErr != nil {
				return wErr
			}
		default:
			return err
		}
	}

	err = step("7. Экспорт итогов")
	if err != nil {
		return err
	}

	for _, format := range []string{"csv", "json"} {
		exp, ok := c.exporters[format]
		if !ok {
			return fmt.Errorf("экспортёр %q не зарегистрирован", format)
		}

		_, err = fmt.Fprintf(c.out, "--- %s ---\n", format)
		if err != nil {
			return err
		}

		err = c.svc.ExportFinal(c.out, exp)
		if err != nil {
			return err
		}
	}
	return nil
}

// demoStudent добавляет студента и печатает результат
func (c *CLI) demoStudent(name, group string) (domain.StudentID, error) {
	id, err := c.svc.AddStudent(name, group)
	if err != nil {
		return id, err
	}

	_, err = fmt.Fprintf(c.out, "  + студент %s: %s (%s)\n", id, name, group)
	if err != nil {
		return id, err
	}

	return id, nil
}

// demoCriterion добавляет критерий и печатает результат
func (c *CLI) demoCriterion(name string, max, weight float64) (domain.CriterionID, error) {
	id, err := c.svc.AddCriterion(name, max, weight)
	if err != nil {
		return id, err
	}

	_, err = fmt.Fprintf(c.out, "  + критерий %s: %s (max %g, вес %g)\n", id, name, max, weight)
	if err != nil {
		return id, err
	}

	return id, nil
}

// expectRejected печатает отказ модели, если ошибки нет или она не та — демо считается проваленным.
func (c *CLI) expectRejected(what string, want, got error) error {
	if !errors.Is(got, want) {
		return fmt.Errorf("демо: «%s» должно быть отклонено с %v, получено: %v", what, want, got)
	}

	_, err := fmt.Fprintf(c.out, "  ✓ отклонено — %s\n      %v\n", what, got)
	if err != nil {
		return err
	}

	return nil
}

package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Ararat25/soda-TAS/internal/domain"
	"github.com/Ararat25/soda-TAS/internal/usecase"
)

// CLI — консольный интерфейс приложения
type CLI struct {
	svc       *usecase.Service
	exporters map[string]usecase.Exporter
	in        *bufio.Reader
	out       io.Writer
}

// New создаёт консольный интерфейс
func New(svc *usecase.Service, exporters map[string]usecase.Exporter) *CLI {
	return &CLI{svc: svc, exporters: exporters, in: bufio.NewReader(os.Stdin), out: os.Stdout}
}

// Run запускает приложение (без аргументов — интерактивное меню, с demo - денстрационный режим)
func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		return c.loop()
	}

	if args[0] == "demo" {
		return c.Demo()
	}

	return fmt.Errorf("неизвестная команда %q (доступно: demo)", args[0])
}

const menu = `
=== ПОС: преподавательская оценка студентов ===
 1. Добавить студента          6. Исправить оценку
 2. Список студентов           7. Отменить оценку
 3. Добавить критерий          8. Оценки студента
 4. Список критериев           9. Итог студента
 5. Выставить оценку          10. Экспорт итогов
 0. Выход

`

// commands возвращает соответствие пунктов меню обработчикам
func (c *CLI) commands() map[string]func() error {
	return map[string]func() error{
		"1":  c.addStudent,
		"2":  c.listStudents,
		"3":  c.addCriterion,
		"4":  c.listCriteria,
		"5":  c.setScore,
		"6":  c.correctScore,
		"7":  c.cancelScore,
		"8":  c.studentEvaluations,
		"9":  c.finalGrade,
		"10": c.exportFinal,
	}
}

// loop запускает интерактивное меню
func (c *CLI) loop() error {
	cmds := c.commands()

	for {
		_, err := fmt.Fprint(c.out, menu)
		if err != nil {
			return err
		}

		choice, err := c.ask("Пункт")
		if err != nil {
			return eofToNil(err)
		}

		if choice == "0" {
			_, err = fmt.Fprintln(c.out, "До свидания!")
			if err != nil {
				return err
			}
			return nil
		}

		handler, ok := cmds[choice]
		if !ok {
			_, err = fmt.Fprintln(c.out, "Нет такого пункта меню.")
			if err != nil {
				return err
			}
			continue
		}

		err = handler()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			c.reportError(err)
		}
	}
}

// eofToNil превращает io.EOF в nil (штатное завершение ввода)
func eofToNil(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// ---------- ввод ----------

// ask запрашивает строку у пользователя
func (c *CLI) ask(prompt string) (string, error) {
	_, err := fmt.Fprintf(c.out, "%s: ", prompt)
	if err != nil {
		return "", err
	}

	line, err := c.in.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// askFloat запрашивает число у пользователя
func (c *CLI) askFloat(prompt string) (float64, error) {
	s, err := c.ask(prompt)
	if err != nil {
		return 0, err
	}

	v, pErr := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
	if pErr != nil {
		return 0, fmt.Errorf("%w: %q — не число", domain.ErrInvalidValue, s)
	}
	return v, nil
}

// ---------- вывод ошибок ----------

// friendly — человекочитаемые описания доменных ошибок
var friendly = []struct {
	err  error
	text string
}{
	{domain.ErrScoreOutOfRange, "балл выходит за пределы критерия"},
	{domain.ErrUnknownCriterion, "такого критерия нет в рубрикаторе"},
	{domain.ErrDuplicateCriterion, "такой критерий уже есть"},
	{domain.ErrAlreadyEvaluated, "оценка уже выставлена, используйте «Исправить оценку»"},
	{domain.ErrNotFound, "не найдено"},
	{domain.ErrNoEvaluations, "у студента ещё нет оценок"},
	{domain.ErrInconsistentEvaluations, "набор оценок несогласован"},
	{domain.ErrInvalidValue, "некорректное значение"},
}

// reportError выводит ошибку пользователю
func (c *CLI) reportError(err error) {
	for _, f := range friendly {
		if errors.Is(err, f.err) {
			_, err = fmt.Fprintf(c.out, "Ошибка: %s\n  (%v)\n", f.text, err)
			if err != nil {
				return
			}

			return
		}
	}
	_, err = fmt.Fprintf(c.out, "Ошибка: %v\n", err)
	if err != nil {
		return
	}
}

// ---------- команды ----------

// addStudent добавляет студента
func (c *CLI) addStudent() error {
	name, err := c.ask("ФИО")
	if err != nil {
		return err
	}

	group, err := c.ask("Группа")
	if err != nil {
		return err
	}

	id, err := c.svc.AddStudent(name, group)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.out, "Студент добавлен, id = %s\n", id)
	if err != nil {
		return err
	}
	return nil
}

// listStudents выводит список студентов
func (c *CLI) listStudents() error {
	students, err := c.svc.ListStudents()
	if err != nil {
		return err
	}

	if len(students) == 0 {
		_, err = fmt.Fprintln(c.out, "Студентов пока нет.")
		if err != nil {
			return err
		}
		return nil
	}

	for _, s := range students {
		_, err = fmt.Fprintf(c.out, "  %-5s %-25s %s\n", s.ID(), s.Name(), s.Group())
		if err != nil {
			return err
		}
	}

	return nil
}

// addCriterion добавляет критерий
func (c *CLI) addCriterion() error {
	name, err := c.ask("Название критерия")
	if err != nil {
		return err
	}

	maxScore, err := c.askFloat("Максимальный балл")
	if err != nil {
		return err
	}

	weight, err := c.askFloat("Вес")
	if err != nil {
		return err
	}

	id, err := c.svc.AddCriterion(name, maxScore, weight)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.out, "Критерий добавлен, id = %s\n", id)
	if err != nil {
		return err
	}
	return nil
}

// listCriteria выводит список критериев
func (c *CLI) listCriteria() error {
	criteria, err := c.svc.ListCriteria()
	if err != nil {
		return err
	}

	if len(criteria) == 0 {
		_, err = fmt.Fprintln(c.out, "Критериев пока нет.")
		if err != nil {
			return err
		}
		return nil
	}

	for _, cr := range criteria {
		_, err = fmt.Fprintf(c.out, "  %-5s %-20s max=%-6g вес=%g\n", cr.ID(), cr.Name(), cr.Max().Value(), cr.Weight().Value())
		if err != nil {
			return err
		}
	}

	return nil
}

// askScoreTarget — общий ввод для выставления и исправления оценки.
func (c *CLI) askScoreTarget() (domain.StudentID, domain.CriterionID, float64, string, error) {
	sid, err := c.ask("id студента")
	if err != nil {
		return "", "", 0, "", err
	}

	cid, err := c.ask("id критерия")
	if err != nil {
		return "", "", 0, "", err
	}

	score, err := c.askFloat("Балл")
	if err != nil {
		return "", "", 0, "", err
	}

	comment, err := c.ask("Комментарий (можно пропустить)")
	if err != nil {
		return "", "", 0, "", err
	}

	return domain.StudentID(sid), domain.CriterionID(cid), score, comment, nil
}

// setScore выставляет оценку
func (c *CLI) setScore() error {
	sid, cid, score, comment, err := c.askScoreTarget()
	if err != nil {
		return err
	}

	if err = c.svc.SetScore(sid, cid, score, comment); err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.out, "Оценка выставлена.")
	if err != nil {
		return err
	}

	return nil
}

// correctScore исправляет оценку
func (c *CLI) correctScore() error {
	sid, cid, score, comment, err := c.askScoreTarget()
	if err != nil {
		return err
	}

	err = c.svc.CorrectScore(sid, cid, score, comment)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.out, "Оценка исправлена.")
	if err != nil {
		return err
	}

	return nil
}

// cancelScore отменяет оценку
func (c *CLI) cancelScore() error {
	sid, err := c.ask("id студента")
	if err != nil {
		return err
	}

	cid, err := c.ask("id критерия")
	if err != nil {
		return err
	}

	err = c.svc.CancelScore(domain.StudentID(sid), domain.CriterionID(cid))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.out, "Оценка отменена.")
	if err != nil {
		return err
	}

	return nil
}

// studentEvaluations выводит оценки студента
func (c *CLI) studentEvaluations() error {
	sid, err := c.ask("id студента")
	if err != nil {
		return err
	}
	return c.printEvaluations(domain.StudentID(sid))
}

// printEvaluations печатает оценки студента
func (c *CLI) printEvaluations(sid domain.StudentID) error {
	evals, err := c.svc.ListEvaluations(sid)
	if err != nil {
		return err
	}

	if len(evals) == 0 {
		_, err = fmt.Fprintln(c.out, "Оценок пока нет.")
		if err != nil {
			return err
		}
		return nil
	}

	criteria, err := c.svc.ListCriteria()
	if err != nil {
		return err
	}

	byID := make(map[domain.CriterionID]domain.Criterion, len(criteria))
	for _, cr := range criteria {
		byID[cr.ID()] = cr
	}

	for _, e := range evals {
		cr := byID[e.CriterionID()]

		_, err = fmt.Fprintf(c.out, "  %-5s %-20s %g из %g  [%s] %s\n", e.CriterionID(), cr.Name(),
			e.Score().Value(), cr.Max().Value(), e.At().Format("02.01.2006 15:04"), e.Comment())
		if err != nil {
			return err
		}
	}

	return nil
}

// finalGrade выводит итоговую оценку студента
func (c *CLI) finalGrade() error {
	sid, err := c.ask("id студента")
	if err != nil {
		return err
	}

	g, err := c.svc.FinalGrade(domain.StudentID(sid))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.out, "Итог: %.1f%%\n", g.Percent())
	if err != nil {
		return err
	}

	return nil
}

// exportFinal выгружает итоги в выбранном формате
func (c *CLI) exportFinal() error {
	formats := make([]string, 0, len(c.exporters))
	for f := range c.exporters {
		formats = append(formats, f)
	}

	sort.Strings(formats)

	format, err := c.ask(fmt.Sprintf("Формат (%s)", strings.Join(formats, "/")))
	if err != nil {
		return err
	}

	exp, ok := c.exporters[strings.ToLower(format)]
	if !ok {
		return fmt.Errorf("%w: неизвестный формат %q", domain.ErrInvalidValue, format)
	}

	path, err := c.ask("Файл (при пропуске — вывести на экран)")
	if err != nil {
		return err
	}
	if path == "" {
		return c.svc.ExportFinal(c.out, exp)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	if err = c.svc.ExportFinal(f, exp); err != nil {
		err = f.Close()
		if err != nil {
			return err
		}

		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.out, "Сохранено в %s\n", path)
	if err != nil {
		return err
	}

	return nil
}

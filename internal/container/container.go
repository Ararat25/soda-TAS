package container

import (
	"fmt"
	"time"

	"github.com/Ararat25/soda-TAS/internal/domain"
	"github.com/Ararat25/soda-TAS/internal/export"
	"github.com/Ararat25/soda-TAS/internal/idgen"
	"github.com/Ararat25/soda-TAS/internal/repository"
	"github.com/Ararat25/soda-TAS/internal/repository/memory"
	"github.com/Ararat25/soda-TAS/internal/transport/cli"
	"github.com/Ararat25/soda-TAS/internal/usecase"

	"go.uber.org/dig"
)

// provider — пара конструктор+опции для регистрации в DI-контейнере
type provider struct {
	constructor any
	opts        []dig.ProvideOption
}

// New собирает DI-контейнер приложения
func New() (*dig.Container, error) {
	c := dig.New()

	providers := []provider{
		// хранение: конкретные in-memory структуры публикуются как порты
		{memory.NewStudentRepo, []dig.ProvideOption{dig.As(new(repository.StudentRepository))}},
		{memory.NewRubricRepo, []dig.ProvideOption{dig.As(new(repository.RubricRepository))}},
		{memory.NewEvaluationRepo, []dig.ProvideOption{dig.As(new(repository.EvaluationRepository))}},

		// инфраструктура приложения
		{idgen.New, []dig.ProvideOption{dig.As(new(usecase.IDGenerator))}},
		{func() usecase.Clock { return usecase.Clock(time.Now) }, nil},

		// домен: правило итога собирается здесь — меняется одной строкой
		{newScoringStrategy, nil},
		{domain.NewGradeCalculator, nil},
		{func(now usecase.Clock) *domain.EvaluationFactory {
			return domain.NewEvaluationFactory(now)
		}, nil},

		// форматы выгрузки
		{export.NewCSV, nil},
		{export.NewJSON, nil},
		{newExporters, nil},

		// приложение и консоль
		{usecase.NewService, nil},
		{cli.New, nil},
	}

	for _, p := range providers {
		err := c.Provide(p.constructor, p.opts...)
		if err != nil {
			return nil, fmt.Errorf("container: %w", err)
		}
	}

	return c, nil
}

// newScoringStrategy - взвешенное среднее с округлением до 0.1
func newScoringStrategy() (domain.ScoringStrategy, error) {
	r, err := domain.NewRounding(domain.WeightedAverage{}, 1)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// newExporters - реестр форматов
func newExporters(csv *export.CSV, js *export.JSON) map[string]usecase.Exporter {
	return map[string]usecase.Exporter{"csv": csv, "json": js}
}

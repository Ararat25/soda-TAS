# ПОС — преподавательская оценка студентов (TAS)

Консольный модуль на Go: студенты, критерии (рубрикатор), оценки, честный итог и выгрузки.

DDD, SOLID, паттерны GoF

## Содержание

1. [Запуск](#1-запуск)
2. [Ключевые решения](#2-ключевые-решения)
3. [Структура проекта](#3-структура-проекта)
4. [Карта DDD](#4-карта-ddd)
5. [SOLID](#5-solid)
6. [Паттерны GoF](#6-паттерны-gof)
7. [DI-контейнер](#7-di-контейнер)
8. [Тесты](#8-тесты)
9. [Демонстрация](#9-демонстрация)
10. [Что сознательно не сделано](#10-что-сознательно-не-сделано)
11. [Роль ИИ](#11-роль-ии)

---

## 1. Запуск

Нужен Go 1.22+. Точка входа: [`cmd/tas/main.go`](cmd/tas/main.go).

```bash
go mod tidy                 # подтянет go.uber.org/dig
go test ./...               # юнит-тесты домена и сценариев
go run ./cmd/tas demo       # сквозной сценарий из ТЗ
go run ./cmd/tas            # интерактивное меню
```

Данные хранятся в памяти и пропадают при выходе. Это осознанное упрощение (см. [раздел 10](#10-что-сознательно-не-сделано)).
Хранилище находится за портом, поэтому его можно заменить без правок домена.

В меню ([`cli.go`](internal/transport/cli/cli.go)) можно добавить и посмотреть студентов и критерии,
выставить, исправить и отменить оценку, посмотреть оценки студента, получить итог
и выгрузить итоги в CSV или JSON (на экран или в файл).

---

## 2. Ключевые решения

Принятые решения:

| Вопрос                                          | Решение                                                                                                                | Почему                                                                   |
|-------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| Единица итога                                   | Проценты `[0, 100]`                                                                                                    | Не зависит от того, сколько баллов в каждом критерии                     |
| Формула по умолчанию                            | `Σ(score / max · weight) / Σweight · 100` ([`scoring.go`](internal/domain/scoring.go))                                 |                                                                          |
| Не все критерии оценены                         | Считаются только оценённые                                                                                             | Неоценённый критерий означает, что оценки ещё нет, и в итог он не входит |
| Нет ни одной оценки                             | Ошибка `ErrNoEvaluations`, в отчёте строка с `graded=false` ([`service.go`](internal/usecase/service.go))              | Итог «0%» не соответствовал бы действительности                          |
| Повторная оценка той же пары (студент+критерий) | Ошибка `ErrAlreadyEvaluated`, править нужно через «Исправить» ([`service.go`](internal/usecase/service.go))            | Идентичность оценки определяется парой, дублей быть не может             |
| Рубрикатор изменился после выставления оценок   | Калькулятор перепроверяет каждую оценку по **текущему** рубрикатору ([`calculator.go`](internal/domain/calculator.go)) | ТЗ требует, чтобы итог был согласован с текущим набором оценок           |
| Откуда берутся ID                               | Генерирует приложение ([`sequence.go`](internal/idgen/sequence.go)) через порт                                         | Генерация — деталь приложения                                            |

---

## 3. Структура проекта

- [`cmd/tas/main.go`](cmd/tas/main.go) — только собирает контейнер и запускает CLI
- `internal/`
    - `domain/` — ядро, ни от чего снаружи не зависит
        - [`errors.go`](internal/domain/errors.go) — доменные ошибки
        - [`values.go`](internal/domain/values.go) — Value Object'ы
        - [`student.go`](internal/domain/student.go) — Entity
        - [`rubric.go`](internal/domain/rubric.go) — Aggregate Root + Entity
        - [`evaluation.go`](internal/domain/evaluation.go) — Entity
        - [`evaluation_factory.go`](internal/domain/evaluation_factory.go) — **[GoF Factory]**
        - [`scoring.go`](internal/domain/scoring.go) — **[GoF Strategy]**: `WeightedAverage`, `SimpleSum`
        - [`scoring_decorators.go`](internal/domain/scoring_decorators.go) — **[GoF Decorator]**: `Rounding`, `Penalty`
        - [`calculator.go`](internal/domain/calculator.go) — Domain Service
        - тесты: [`helpers_test.go`](internal/domain/helpers_test.go), [`values_test.go`](internal/domain/values_test.go), [`rubric_test.go`](internal/domain/rubric_test.go), [`evaluation_factory_test.go`](internal/domain/evaluation_factory_test.go), [`calculator_test.go`](internal/domain/calculator_test.go)
    - `repository/`
        - [`repository.go`](internal/repository/repository.go) — **порты** хранения
        - `memory/` — in-memory адаптеры: [`student.go`](internal/repository/memory/student.go), [`rubric.go`](internal/repository/memory/rubric.go), [`evaluation.go`](internal/repository/memory/evaluation.go)
    - `usecase/`
        - [`service.go`](internal/usecase/service.go) — **[GoF Facade]**, все сценарии приложения и порты `IDGenerator` и `Clock`
        - [`exporter.go`](internal/usecase/exporter.go) — порт `Exporter`
        - [`service_test.go`](internal/usecase/service_test.go) — тесты сценариев
    - `export/` — адаптеры `Exporter`: [`csv.go`](internal/export/csv.go), [`json.go`](internal/export/json.go)
    - [`idgen/sequence.go`](internal/idgen/sequence.go) — адаптер `IDGenerator`
    - `transport/cli/` — консоль: [`cli.go`](internal/transport/cli/cli.go) (меню), [`demo.go`](internal/transport/cli/demo.go) (сквозной сценарий)
    - [`container/container.go`](internal/container/container.go) — DI-контейнер (uber/dig)

Направление зависимостей — всегда внутрь, к домену:

```
transport/cli ──► usecase ──► repository (порты) ──► domain
                     ▲                 ▲
                export, idgen     repository/memory
              (адаптеры портов)   (адаптеры портов)
```

Пакет `domain` не импортирует ничего из проекта.

Вся доменная модель лежит в одном пакете `domain`, без отдельных пакетов на каждый агрегат.
В Go пакет должен быть цельным по смыслу и не обязан повторять диаграмму классов.
`GradeCalculator` работает и с рубрикатором, и с оценками, а фабрика оценки обращается к рубрикатору.
Если разнести эти типы по подпакетам, быстро появятся циклические импорты и повторяющиеся имена вроде `rubric.Rubric`.
Границы агрегатов обозначены файлами и описаны ниже.
Недостаток такого решения (инкапсуляция внутри пакета слабее, чем между пакетами) разобран в [разделе 5](#где-упростили-или-почти-нарушили).

---

## 4. Карта DDD

### 4.1. Ubiquitous Language

Код использует те же слова, что и ТЗ.

| Термин ТЗ          | В коде                                                    | Тип                             |
|--------------------|-----------------------------------------------------------|---------------------------------|
| студент            | [`domain.Student`](internal/domain/student.go)            | Entity                          |
| критерий           | [`domain.Criterion`](internal/domain/rubric.go)           | Entity внутри агрегата `Rubric` |
| рубрикатор         | [`domain.Rubric`](internal/domain/rubric.go)              | Aggregate Root                  |
| оценка             | [`domain.Evaluation`](internal/domain/evaluation.go)      | Entity                          |
| балл               | [`domain.Score`](internal/domain/values.go)               | Value Object                    |
| максимум критерия  | [`domain.MaxScore`](internal/domain/values.go)            | Value Object                    |
| вклад в итог (вес) | [`domain.Weight`](internal/domain/values.go)              | Value Object                    |
| итог               | [`domain.FinalGrade`](internal/domain/values.go)          | Value Object (вычисляется)      |
| расчёт итога       | [`domain.GradeCalculator`](internal/domain/calculator.go) | Domain Service                  |

### 4.2. Entity и Value Object — осознанное разделение

**Пример: `Student` и `Score`.**

- [`Student`](internal/domain/student.go) — **Entity**. У студента есть идентичность (`StudentID`). Если сменить фамилию или группу,
  это тот же студент. Два студента с одинаковым именем и группой — разные люди.
- [`Score`](internal/domain/values.go) — **Value Object**. У балла нет собственной идентичности, «7» всегда равно «7». Он неизменяем
  (поля неэкспортируемые, методов-мутаторов нет), сравнивается по значению и создаётся только
  через конструктор `NewScore`, который не принимает `NaN`, `±Inf` и отрицательные значения.

Остальные типы разделены так же. [`Criterion`](internal/domain/rubric.go) и [`Evaluation`](internal/domain/evaluation.go) — сущности, у них есть идентичность
и жизненный цикл. `MaxScore`, `Weight` и `FinalGrade` — значения ([`values.go`](internal/domain/values.go)).
`Evaluation` идентифицируется парой (студент, критерий) и отдельного ID не имеет.
Вторая оценка по той же паре считается исправлением первой.

### 4.3. Агрегаты и их границы

| Агрегат                                                       | Корень       | Что внутри                  | Почему такая граница                                                                                                       |
|---------------------------------------------------------------|--------------|-----------------------------|----------------------------------------------------------------------------------------------------------------------------|
| Student ([`student.go`](internal/domain/student.go))          | `Student`    | ФИО, группа                 | Живёт независимо от рубрикатора и оценок                                                                                   |
| Rubric ([`rubric.go`](internal/domain/rubric.go))             | `Rubric`     | `Criterion`                 | Критерии не существуют вне рубрикатора, уникальность названий охраняет корень                                              |
| Evaluation ([`evaluation.go`](internal/domain/evaluation.go)) | `Evaluation` | `Score`, комментарий, время | Оценки меняются часто и их много. Если бы они находились внутри `Rubric`, любое исправление блокировало бы весь рубрикатор |

Агрегаты ссылаются друг на друга **по идентификатору** (`StudentID`, `CriterionID`).
Фабрика и `Correct` получают `*Rubric` только для проверки и не сохраняют его в оценке.

### 4.4. Инварианты и где они охраняются

| Правило                                                     | Где проверяется                                                                                                                                                                                                                                                          |
|-------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Балл не выходит за пределы критерия                         | `Criterion.Validate` ([`rubric.go`](internal/domain/rubric.go)), вызывается из [`EvaluationFactory.Create`](internal/domain/evaluation_factory.go), [`Evaluation.Correct`](internal/domain/evaluation.go) и [`GradeCalculator.Calculate`](internal/domain/calculator.go) |
| Оценить можно только критерий из рубрикатора                | `Rubric.Criterion(id)` → `ErrUnknownCriterion` ([`rubric.go`](internal/domain/rubric.go)), вызывается из фабрики, `Correct` и калькулятора                                                                                                                               |
| Балл ≥ 0, максимум > 0, вес > 0, итог ∈ [0, 100]            | Конструкторы Value Object'ов ([`values.go`](internal/domain/values.go))                                                                                                                                                                                                  |
| Критерии уникальны по ID и по названию (без учёта регистра) | `Rubric.AddCriterion` ([`rubric.go`](internal/domain/rubric.go))                                                                                                                                                                                                         |
| Итог согласован с текущими оценками                         | `GradeCalculator.Calculate` ([`calculator.go`](internal/domain/calculator.go))                                                                                                                                                                                           |
| Критерий нельзя создать из «нулевых» VO                     | `NewCriterion` ([`rubric.go`](internal/domain/rubric.go), см. ниже)                                                                                                                                                                                                      |

Все правила находятся в домене. Консоль и сценарии только превращают
ввод в Value Object'ы и показывают ошибки. Благодаря этому `Evaluation` остаётся полноценной сущностью
с поведением (`Correct`) и защищённым состоянием.

**Нулевое значение в Go.** Из другого пакета нельзя записать `-5` в `MaxScore`, однако
`var m domain.MaxScore` даёт значение `v == 0` в обход конструктора. Поэтому корень агрегата
повторно проверяет переданные Value Object'ы. Для этого `NewCriterion` в [`rubric.go`](internal/domain/rubric.go) отклоняет нулевые `MaxScore` и `Weight`.
Без этой проверки критерий с нулевым максимумом попал бы в рубрикатор, а расчёт итога упал бы с делением на ноль далеко от места ошибки.

### 4.5. Repository (порт)

Порты объявлены в [`internal/repository/repository.go`](internal/repository/repository.go). Это `StudentRepository`, `RubricRepository`
и `EvaluationRepository`. Они оперируют доменными типами («сохрани оценку», «найди оценки студента»)
и ничего не говорят о способе хранения. Реализации лежат в `repository/memory`
([`student.go`](internal/repository/memory/student.go), [`rubric.go`](internal/repository/memory/rubric.go), [`evaluation.go`](internal/repository/memory/evaluation.go)).
Ошибка «не найдено» определена в домене (`domain.ErrNotFound` в [`errors.go`](internal/domain/errors.go)), поэтому не зависит от конкретного хранилища.

### 4.6. Domain Service

Расчёт итога ([`GradeCalculator.Calculate(rubric, evaluations)`](internal/domain/calculator.go)) затрагивает и рубрикатор, и оценки,
поэтому не принадлежит ни одной сущности и оформлен как Domain Service. Сначала сервис
проверяет, что набор оценок допустим, и только после этого передаёт его стратегии, которая считает итог по формуле.
Сервис отклоняет:

- пустой набор (`ErrNoEvaluations`);
- оценки разных студентов (`ErrInconsistentEvaluations`);
- две оценки по одному критерию (`ErrInconsistentEvaluations`);
- оценку по критерию не из рубрикатора (`ErrUnknownCriterion`);
- «устаревшую» оценку, если максимум критерия уменьшили после выставления (`ErrScoreOutOfRange`).

Загрузкой данных занимается `usecase` ([`service.go`](internal/usecase/service.go)), поэтому сервис получает готовые объекты и остаётся чистым.

---

## 5. SOLID

### Где принципы проявились

| Принцип | Где                              | Как                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
|---------|----------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **S**   | Слои                             | `domain` — правила, [`usecase`](internal/usecase/service.go) — оркестрация, [`transport/cli`](internal/transport/cli/cli.go) — ввод-вывод, [`export`](internal/export/csv.go) — форматы, [`repository/memory`](internal/repository/memory/evaluation.go) — хранение.                                                                                                                                                                                                                                                                                                      |
| **O**   | Стратегия итога, формат выгрузки | Новое правило итога — новый тип с методом `Grade` ([`scoring.go`](internal/domain/scoring.go)) и одна строка в [`container.go`](internal/container/container.go). Новый формат выгрузки — новый тип в [`export`](internal/export/json.go) и одна строка в реестре форматов. Ни домен, ни `usecase`, ни консоль не меняются                                                                                                                                                                                                                                                |
| **L**   | `ScoringStrategy` и декораторы   | Любая стратегия и любой декоратор ([`scoring.go`](internal/domain/scoring.go), [`scoring_decorators.go`](internal/domain/scoring_decorators.go)) гарантируют одно и то же. При корректных данных они возвращают `FinalGrade ∈ [0, 100]`, не меняют аргументы и учитывают только оценённые критерии. Это проверяет `TestStrategies_HonourContract` в [`calculator_test.go`](internal/domain/calculator_test.go). Каждая стратегия и композиция даёт 100 на максимальных баллах и 0 на нулевых. Декоратор можно подставить вместо базовой стратегии, и контракт не ломается |
| **I**   | Порты                            | Порты маленькие и разделены по назначению. Хранилища разнесены по агрегатам ([`repository.go`](internal/repository/repository.go)), у `Exporter` один метод ([`exporter.go`](internal/usecase/exporter.go)), у `IDGenerator` и `Clock` тоже по одному ([`service.go`](internal/usecase/service.go)). «IGodRepository» в проекте нет                                                                                                                                                                                                                                       |
| **D**   | Всё приложение                   | `usecase` зависит от интерфейсов (`repository.*`, `Exporter`, `IDGenerator`), реализации подставляет DI-контейнер на краю приложения ([`container.go`](internal/container/container.go))                                                                                                                                                                                                                                                                                                                                                                                  |

### Где упростили или почти нарушили

1. **`usecase.Service` — один тип на все сценарии** ([`service.go`](internal/usecase/service.go)). Формально у него много причин для изменений
   (студенты, рубрикатор, оценки, отчёты). Это сознательное упрощение, ведь сценариев немного, а `Service`
   одновременно выполняет роль Facade для консоли. Если модуль вырастет (пересдачи, несколько курсов), его стоит
   разделить по агрегатам.
2. **ISP не доведён до конца в хранилищах** ([`repository.go`](internal/repository/repository.go)). Порты разнесены по агрегатам, но не делятся на чтение
   и запись. `Service` зависит от `EvaluationRepository` целиком, хотя `ListEvaluations`
   использует только `ByStudent`. Для четырёх методов дробление добавило бы интерфейсов без заметной пользы.
3. **Инкапсуляция внутри `domain` держится на дисциплине.** Неэкспортируемые поля скрыты от других
   пакетов, но любой файл внутри `domain` может писать в них напрямую. Поэтому объекты
   создаются только через конструкторы и фабрику ([`evaluation_factory.go`](internal/domain/evaluation_factory.go)), а `Rubric.Criteria()` ([`rubric.go`](internal/domain/rubric.go))
   возвращает копию внутреннего слайса.
4. **In-memory `RubricRepo` отдаёт один и тот же указатель** ([`rubric.go`](internal/repository/memory/rubric.go)). С настоящей БД так не будет,
   поэтому `usecase` всё равно вызывает `Save` после изменения.
5. **Фабрика зависит от конкретного рубрикатора.** [`EvaluationFactory.Create`](internal/domain/evaluation_factory.go) принимает `*Rubric`.
   Узкий интерфейс `CriterionLookup` освободил бы фабрику от этой зависимости, но
   ради единственной реализации он стал бы декоративной абстракцией.

---

## 6. Паттерны GoF

Выбрано по одному паттерну на каждую группу и дополнительно Facade во второй группе. Всего получилось четыре паттерна. В коде каждое место помечено комментарием `[GoF ...]`, по нему его можно найти поиском по проекту.

| Группа        | Паттерн       | Где                                                                                    | Какую боль закрывает                                                                                                                                                                                                                  |
|---------------|---------------|----------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Порождающий   | **Factory**   | [`domain.EvaluationFactory.Create`](internal/domain/evaluation_factory.go)             | Фабрика — единственная точка создания оценки, поэтому создать оценку в обход правил нельзя. Она всегда проверяет, что критерий есть в рубрикаторе и `0 ≤ score ≤ max`. Время передаётся извне, поэтому тесты детерминированы          |
| Структурный   | **Decorator** | [`domain.Rounding`, `domain.Penalty`](internal/domain/scoring_decorators.go)           | Модификаторы итога (округление, штраф) собраны в одном месте и не попадают в стратегии и UI. Декоратор сам реализует `ScoringStrategy`, поэтому их можно комбинировать без правки ядра, например `Rounding(Penalty(WeightedAverage))` |
| Структурный   | **Facade**    | [`usecase.Service`](internal/usecase/service.go)                                       | Консоль ничего не знает о репозиториях, фабрике и калькуляторе, а сценарии собраны в одном месте и не размазаны по UI                                                                                                                 |
| Поведенческий | **Strategy**  | [`domain.ScoringStrategy`: `WeightedAverage`, `SimpleSum`](internal/domain/scoring.go) | Если заказчик захочет другое правило итога, достаточно изменить одну строку в [`container.go`](internal/container/container.go), калькулятор править не нужно                                                                         |

Паттерны работают вместе. [`GradeCalculator`](internal/domain/calculator.go) (Domain Service) принимает `ScoringStrategy`,
стратегию можно обернуть в декораторы, а оценки, с которыми она работает, создаёт фабрика.

Примеры из ТЗ, которые не реализованы:

- **Builder (сборка рубрикатора).** Рубрикатор собирается несколькими вызовами `AddCriterion`, сложной
  пошаговой сборки с обязательными шагами нет, поэтому Builder был бы декоративным.
- **Template Method (импорт ведомостей).** Импорт не входит в обязательные сценарии.

---

## 7. DI-контейнер

Используется `go.uber.org/dig`. Все зависимости регистрируются в одном месте, [`internal/container/container.go`](internal/container/container.go):

- конкретные in-memory репозитории публикуются как порты через `dig.As(new(repository.XxxRepository))`;
- правило итога собирает функция `newScoringStrategy`, она возвращает `Rounding(WeightedAverage, 1)`;
- форматы выгрузки ([`csv.go`](internal/export/csv.go), [`json.go`](internal/export/json.go)) собираются в реестр `map[string]usecase.Exporter`;
- [`main.go`](cmd/tas/main.go) только создаёт контейнер и вызывает `Invoke`.

Остальные пакеты ничего не знают о контейнере и получают зависимости через конструкторы.
Порядок создания объектов контейнер определяет сам по сигнатурам конструкторов, поэтому ручной сборки в `main` нет.

---

## 8. Тесты

```bash
go test ./...
```

Тесты табличные

| Файл                                                                       | Что проверяет                                                                                                                                  |
|----------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| [`values_test.go`](internal/domain/values_test.go)                         | Конструкторы VO, значения `NaN` и `Inf`, отрицательные числа, границы итога                                                                    |
| [`rubric_test.go`](internal/domain/rubric_test.go)                         | Дубли критериев (по ID и по названию без учёта регистра), «нулевые» VO, копия из `Criteria()`                                                  |
| [`evaluation_factory_test.go`](internal/domain/evaluation_factory_test.go) | 12 из 10, граница 0 и max, неизвестный критерий; `Correct` не меняет состояние при отказе                                                      |
| [`calculator_test.go`](internal/domain/calculator_test.go)                 | Итог для известного набора по каждой стратегии и композиции декораторов, контракт стратегий (LSP), несогласованные наборы, «устаревшая» оценка |
| [`service_test.go`](internal/usecase/service_test.go)                      | Инварианты на уровне сценариев с реальными in-memory репозиториями, жизненный цикл оценки, итог, экспорт со студентом без оценок               |

Общие вспомогательные функции тестов домена лежат в [`helpers_test.go`](internal/domain/helpers_test.go).

---

## 9. Демонстрация

```bash
go run ./cmd/tas demo
```

Сценарий описан в [`demo.go`](internal/transport/cli/demo.go) и повторяет п.4 функциональных требований ТЗ:

1. завести трёх студентов;
2. завести три критерия (Теория, Практика, Проект);
3. выставить оценки;
4. показать, что модель отвергает «12 из 10», критерий не из рубрикатора и повторную оценку той же пары;
5. исправить и отменить оценки;
6. получить итоги (у студента без оценок выводится «нет оценок»);
7. выгрузить итоги в CSV и в JSON.

Если модель принимает то, что должна отклонить, демо завершается с ошибкой.

```text
=== 1. Студенты ===
  + студент s1: Иван Петров (Б01-001)
  + студент s2: Анна Смирнова (Б01-002)
  + студент s3: Пётр Сидоров (Б01-003)

=== 2. Рубрикатор ===
  + критерий c1: Теория (max 10, вес 1)
  + критерий c2: Практика (max 20, вес 3)
  + критерий c3: Проект (max 30, вес 2)

=== 3. Выставление оценок ===
  + s1 / c1 = 7
  + s1 / c2 = 15
  + s1 / c3 = 24
  + s2 / c1 = 9
  + s2 / c2 = 18

=== 4. Модель не даёт нарушить правила домена ===
  ✓ отклонено — 12 баллов из 10
      score is outside [0, max_score]: got 12, max 10 (criterion "Теория")
  ✓ отклонено — критерий не из рубрикатора
      criterion is not in the rubric: "c99"
  ✓ отклонено — повторная оценка той же пары
      criterion is already evaluated for this student: student "s2", criterion "c1" (use correct)

=== 5. Исправление и отмена ===
  исправлено: s1 / c1 → 8
  отменено:   s2 / c2
  выставлено заново: s2 / c2 = 17
Оценки Ивана:
  c1    Теория               8 из 10  [22.11.2026 22:48] пересмотрено
  c2    Практика             15 из 20  [22.11.2026 22:48] 
  c3    Проект               24 из 30  [22.11.2026 22:48] 

=== 6. Итоги ===
  Иван Петров        77.5%
  Анна Смирнова      86.3%
  Пётр Сидоров       нет оценок

=== 7. Экспорт итогов ===
--- csv ---
﻿student_id,name,group,grade_percent
s1,Иван Петров,Б01-001,77.50
s2,Анна Смирнова,Б01-002,86.30
s3,Пётр Сидоров,Б01-003,
--- json ---
[
  {
    "student_id": "s1",
    "name": "Иван Петров",
    "group": "Б01-001",
    "graded": true,
    "grade_percent": 77.5
  },
  {
    "student_id": "s2",
    "name": "Анна Смирнова",
    "group": "Б01-002",
    "graded": true,
    "grade_percent": 86.3
  },
  {
    "student_id": "s3",
    "name": "Пётр Сидоров",
    "group": "Б01-003",
    "graded": false,
    "grade_percent": null
  }
]

```

---

## 10. Что сознательно не сделано

| Что                                             | Почему допустимо                                                                                                                                                                                                 |
|-------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Удаление и редактирование студентов и критериев | По ТЗ изменение и удаление нужны «по необходимости, не самоцель».                                                                                                                                                |
| Постоянное хранилище                            | Хранение стоит за портом ([`repository.go`](internal/repository/repository.go)), сейчас достаточно in-memory                                                                                                     |
| Импорт ведомостей                               | Не входит в обязательные сценарии                                                                                                                                                                                |
| Пересдачи и несколько курсов                    | В ТЗ они отнесены к будущему, сейчас достаточно ядра                                                                                                                                                             |
| Потокобезопасность консоли                      | Репозитории защищены мьютексами, но приложение однопоточное                                                                                                                                                      |

---

## 11. Роль ИИ

При работе над проектом использовался ИИ-ассистент. Он применялся для:

- обсуждения структуры проекта
- подготовки каркаса кода доменного слоя
- составления README.
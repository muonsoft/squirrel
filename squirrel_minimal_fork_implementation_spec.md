# Техническое задание: минимальный поддерживаемый fork Squirrel для PostgreSQL

**Статус:** ready for implementation
**Тип задачи:** fork / refactoring / dependency cleanup / compatibility / regression hardening
**Основной язык:** Go
**Исходные проекты:**
- Masterminds/squirrel: https://github.com/Masterminds/squirrel
- n-r-w/squirrel: https://github.com/n-r-w/squirrel

---

## 1. Контекст

В существующих Go-проектах используется `github.com/Masterminds/squirrel` как SQL builder.

Основные причины отказаться от прямой зависимости от оригинального проекта:

1. `Masterminds/squirrel` фактически находится в режиме минимального сопровождения.
2. Последний release оригинального проекта — `v1.5.4`.
3. В сложных композициях запросов исторически возникали ошибки с нумерацией PostgreSQL placeholders (`$1`, `$2`, ...), особенно во вложенных запросах.
4. Есть полезный активно доработанный fork `github.com/n-r-w/squirrel`, содержащий:
   - исправления nested-query placeholder handling;
   - CTE / recursive CTE;
   - PostgreSQL DML CTE;
   - `UPDATE ... FROM`;
   - расширенную поддержку subquery;
   - ряд дополнительных expression helpers.
5. При этом текущий `n-r-w/squirrel` содержит слишком широкий `go.mod`: `pgx`, `scany`, `testdock`, Docker и множество транзитивных зависимостей. Большая часть этих зависимостей относится к integration tests и не нужна самому SQL builder.
6. Цель — получить маленькую, предсказуемую и самостоятельно сопровождаемую библиотеку, а не новый SQL framework.

---

# 2. Цель

Создать внутренний fork Squirrel, ориентированный на:

- построение SQL;
- PostgreSQL;
- сложные/nested queries;
- корректную работу placeholders;
- CTE;
- минимальный dependency graph;
- высокую совместимость с привычным API Masterminds/squirrel;
- простое долгосрочное сопровождение.

Библиотека должна оставаться **pure SQL builder**.

Она НЕ должна:

- выполнять SQL;
- управлять подключениями;
- сканировать строки;
- содержать pagination/business-policy abstractions;
- зависеть от PostgreSQL driver в основном Go-модуле;
- зависеть от Docker/testcontainers/testdock в основном Go-модуле.

---

# 3. Baseline

В качестве исходной кодовой базы использовать:

```text
github.com/n-r-w/squirrel
tag: v1.6.0
release commit: 4c87dbb
```

Причина выбора `n-r-w`, а не прямого fork `Masterminds`:

- в `n-r-w` уже реализована важная архитектурная правка nested placeholder handling;
- release `v1.4.3` явно исправлял проблемы с нумерацией параметров во вложенных запросах;
- в коде есть разделение между:
  - `Sqlizer.ToSql()`;
  - внутренним `rawSqlizer.toSqlRaw()`;
  - `nestedToSql()`;
- финальная замена `?` в `$N` выполняется после сборки вложенной структуры запроса;
- уже реализованы CTE и PostgreSQL-oriented возможности.

Не начинать с `Masterminds v1.5.4` с последующим ручным cherry-pick десятков изменений, если в ходе анализа не обнаружится критическая причина изменить baseline.

---

# 4. Главный архитектурный принцип

## 4.1. Scope библиотеки

Библиотека отвечает только за:

```text
Go DSL / expressions
        ↓
SQL AST-like composition
        ↓
SQL string + []any
```

Финальный контракт:

```go
type Sqlizer interface {
    ToSql() (string, []any, error)
}
```

Любая функциональность, которая относится к:

```text
database/sql
pgx
connection pool
transactions
rows scanning
pagination policy
search policy
application filtering
repository abstraction
```

не относится к core library.

---

# 5. Требования к совместимости

## 5.1. Приоритет совместимости

Приоритеты:

1. корректность SQL;
2. совместимость с типичным API `Masterminds/squirrel`;
3. минимальный dependency graph;
4. PostgreSQL extensions;
5. дополнительные convenience helpers.

Не сохранять helper только ради API compatibility с `n-r-w`, если helper размывает ответственность библиотеки.

## 5.2. Import compatibility

Package name должен остаться:

```go
package squirrel
```

Module path должен соответствовать новому корпоративному repository.

Если repository/module path уже определён окружением — использовать существующий.

Не придумывать новый публичный module path без необходимости.

## 5.3. Masterminds compatibility

Сохранить привычные core primitives, насколько это возможно:

- `Select`
- `Insert`
- `Update`
- `Delete`
- `StatementBuilder`
- `PlaceholderFormat`
- `Question`
- `Dollar`
- `Colon`
- `AtP`
- `Expr`
- `Eq`
- `NotEq`
- `Lt`
- `LtOrEq`
- `Gt`
- `GtOrEq`
- `Like`
- `ILike`
- `NotLike`
- `NotILike`
- `And`
- `Or`
- `Case`
- `Prefix`
- `Suffix`
- `Join`
- `JoinClause`
- `Where`
- `Having`
- `OrderBy`
- `Limit`
- `Offset`
- `FromSelect`
- `Column`
- `Columns`
- `Values`
- `Set`
- `SetMap`
- `MustSql`
- `DebugSqlizer`

Если API `n-r-w` изменил поведение существующего Masterminds API, необходимо отдельно проверить необходимость изменения.

Особенно проверить `Case`, так как `n-r-w` документирует breaking changes в обработке строковых и числовых значений.

### Требование

Не принимать breaking behavior `n-r-w` автоматически.

Для каждого такого изменения:

1. сравнить Masterminds v1.5.4;
2. сравнить n-r-w v1.6.0;
3. оценить реальную пользу;
4. по умолчанию сохранить наиболее совместимое с Masterminds поведение;
5. добавить regression tests.

---

# 6. Функциональность, которую необходимо сохранить из n-r-w

## 6.1. Nested query placeholder handling — MUST HAVE

Это ключевое изменение fork.

Сохранить архитектуру:

```go
type rawSqlizer interface {
    toSqlRaw() (string, []any, error)
}
```

и механизм:

```go
func nestedToSql(s Sqlizer) (...) {
    if raw, ok := s.(rawSqlizer); ok {
        return raw.toSqlRaw()
    }

    return s.ToSql()
}
```

### Основной invariant

Внутренние builders должны строить запрос с логическими `?` placeholders.

Преобразование:

```text
? → $1, $2, $3...
```

для `Dollar` должно происходить **только после окончательной композиции запроса**.

Не выполнять независимую нумерацию nested query и parent query.

### Пример ожидаемой логики

Концептуально:

```go
sub := Select("id").
    From("accounts").
    Where("tenant_id = ?", tenantID)

q := Select("*").
    From("users").
    Where("account_id IN (?)", sub).
    Where("status = ?", status).
    PlaceholderFormat(Dollar)
```

Результат должен иметь сквозную нумерацию:

```sql
... tenant_id = $1 ... status = $2
```

а `args` должны строго соответствовать порядку placeholders.

---

# 7. PostgreSQL extensions — MUST HAVE

## 7.1. CTE

Сохранить:

- `WITH`;
- несколько CTE;
- `WITH RECURSIVE`;
- CTE body как `Sqlizer`;
- final statement после CTE.

Поддержать:

```sql
WITH a AS (...),
     b AS (...)
SELECT ...
```

и:

```sql
WITH RECURSIVE ...
```

## 7.2. DML CTE

CTE body должен поддерживать как минимум:

- `SELECT`;
- `INSERT`;
- `UPDATE`;
- `DELETE`.

Пример обязательного сценария:

```sql
WITH candidate AS (
    SELECT ...
    FOR UPDATE SKIP LOCKED
),
updated AS (
    UPDATE ...
    FROM candidate
    RETURNING ...
)
SELECT ...
FROM updated
```

Для PostgreSQL этот паттерн является важным production use case.

## 7.3. UPDATE ... FROM

Сохранить поддержку:

```sql
UPDATE table
SET ...
FROM ...
WHERE ...
```

Она должна нормально композироваться с:

- subquery;
- CTE;
- placeholders;
- `RETURNING` через suffix или нативный существующий API, если он имеется.

## 7.4. Subquery expressions

Сохранить возможность использовать subquery как `Sqlizer` в:

- `WHERE`;
- `Eq`;
- `And` / `Or`;
- `Column`;
- `FROM`;
- `JOIN`;
- comparisons;
- CTE;
- expressions.

## 7.5. EXISTS / NOT EXISTS

Если реализация `n-r-w` уже содержит:

- `Exists`;
- `NotExists`;

сохранить их.

Они являются generic SQL primitives и соответствуют scope builder.

## 7.6. Scalar subquery comparisons

Сохранить helpers:

- `Equal`
- `NotEqual`
- `Greater`
- `GreaterOrEqual`
- `Less`
- `LessOrEqual`

при условии, что они корректно используют nested placeholder handling.

## 7.7. NOT

Сохранить generic `Not(Sqlizer)`.

Проверить поведение double-NOT.

## 7.8. COALESCE

Сохранить `Coalesce`, если implementation остаётся маленьким и dependency-free.

Обязательно покрыть nested subqueries и placeholders.

## 7.9. Aggregate expressions

Допускается сохранить:

- `Sum`
- `Count`
- `Avg`
- `Min`
- `Max`

если:

- implementation не вводит зависимостей;
- API очевиден;
- expressions корректно работают с nested builders.

Это secondary priority.

---

# 8. IN / NOT IN

В `n-r-w` helper `In` имеет PostgreSQL-specific оптимизацию:

для нескольких значений может строиться конструкция вида:

```sql
column = ANY(?)
```

а `NotIn` — эквивалент через `ALL`.

Эту функциональность разрешается сохранить, так как fork ориентирован преимущественно на PostgreSQL.

Но требуется:

1. явно задокументировать это поведение;
2. проверить выполнение через `pgx`;
3. проверить:
   - scalar;
   - slice length 0;
   - slice length 1;
   - slice length > 1;
   - subquery;
   - typed slice;
   - `[]uuid.UUID`, если такой тип доступен только в integration test consumer, а не в core;
4. не вводить зависимость от `pgx` в core module.

Если поведение выглядит слишком неочевидным для метода `In`, агент должен зафиксировать это в итоговом отчёте как architectural concern, но не должен самостоятельно полностью передизайнивать API без необходимости.

---

# 9. Что необходимо удалить из core fork

Следующие функции из `n-r-w` не являются частью минимального SQL builder и по умолчанию должны быть удалены.

## 9.1. Search

Удалить:

```go
Select(...).Search(...)
```

Причина:

- задаёт application-level search semantics;
- автоматически приводит колонки к `::text`;
- автоматически выбирает `LIKE`;
- автоматически добавляет `%value%`;
- является PostgreSQL-specific policy, а не SQL primitive.

Пользователь библиотеки может выразить это через обычные expressions.

## 9.2. Pagination abstractions

Удалить:

- `Paginator`;
- `PaginatorType`;
- `PaginatorByPage`;
- `PaginatorByID`;
- `Paginate`;
- `PaginateByPage`;
- `PaginateByID`;
- `SetIDColumn`;
- связанные поля в `selectData`.

Причина:

pagination strategy относится к application/repository layer.

Core builder уже содержит:

```go
Limit(...)
Offset(...)
Where(...)
OrderBy(...)
```

этого достаточно.

## 9.3. OrderByCond

Удалить:

- `OrderCond`;
- `OrderByCond`;
- `OrderByCondOption`;
- mapping numeric column ID → SQL column;
- duplicate filtering logic.

Причина:

это application mapping abstraction.

Дополнительная выгода: после удаления исчезает реальная потребность в:

```text
golang.org/x/exp/slices
```

## 9.4. EqNotEmpty

Удалить:

```go
EqNotEmpty
```

Причина:

семантика "zero value означает отсутствующий фильтр" является application policy и потенциально неожиданна.

Например:

```go
0
""
false
```

могут быть валидными SQL filter values.

Builder не должен самостоятельно интерпретировать их как "filter omitted".

## 9.5. Select alias state helper

Проверить дополнительные alias helpers `n-r-w`, которые автоматически модифицируют списки columns/group/order.

Если alias является stateful/application convenience abstraction — удалить.

Оставить generic expression helper:

```go
Alias(Sqlizer, "alias")
```

если он просто генерирует:

```sql
(expression) AS alias
```

---

# 10. Database execution API

Из `n-r-w` уже удалены DB interaction methods.

Это решение сохранить.

В core library НЕ должно быть:

- `RunWith`;
- `Exec`;
- `ExecContext`;
- `Query`;
- `QueryContext`;
- `QueryRow`;
- `QueryRowContext`;
- `StmtCache`;
- `NewStmtCache`;
- wrappers над `database/sql`;
- wrappers над `pgx`.

`database/sql/driver` допустим только в тех местах, где он нужен для определения SQL value types и не создаёт runtime DB integration.

---

# 11. Dependency policy

## 11.1. Core module

Целевой production dependency graph:

```text
<internal-squirrel>
└── github.com/lann/builder
    └── github.com/lann/ps
```

В идеальном результате root `go.mod` должен содержать только:

```text
github.com/lann/builder
```

и минимально необходимые indirect dependencies.

## 11.2. Запрещённые зависимости в core module

В основном `go.mod` не должно быть:

```text
github.com/jackc/pgx
github.com/georgysavva/scany
github.com/n-r-w/testdock
github.com/ory/dockertest
github.com/docker/*
github.com/golang-migrate/*
github.com/pressly/goose
github.com/stretchr/testify
golang.org/x/exp
```

Если после `go mod tidy` они появляются — определить источник и устранить.

## 11.3. Test dependencies

Unit tests core package по возможности писать на стандартном:

```go
testing
```

Не использовать `testify`, если это не даёт существенной выгоды.

Цель — root module должен оставаться минимальным даже после `go mod tidy`.

## 11.4. lann/builder

`github.com/lann/builder` оставить на первом этапе.

Не переписывать внутреннюю immutable-builder механику в рамках этой задачи.

Причина:

- это фундаментальная часть архитектуры Squirrel;
- её удаление значительно увеличит diff;
- увеличится вероятность API/behavior regression;
- dependency маленькая.

Отдельно зафиксировать технический долг:

> в будущем можно рассмотреть замену `lann/builder` на собственные typed immutable structs, но только отдельной задачей после стабилизации fork.

---

# 12. Go version

Не наследовать требование Go 1.26 только потому, что оно указано в `n-r-w v1.6.0`.

Core SQL builder не должен искусственно повышать минимальную версию Go.

Агент должен:

1. проверить используемые language/library features;
2. определить минимально разумную версию;
3. учитывать corporate/toolchain policy, если она присутствует в repository;
4. не использовать unsupported Go release ради искусственного снижения версии;
5. прогнать CI минимум:
   - на declared minimum Go;
   - на актуальном поддерживаемом Go toolchain.

Если repository не содержит корпоративной политики, выбрать поддерживаемую на момент реализации версию Go, совместимую с используемой инфраструктурой.

Главное требование: повышение Go version должно быть обосновано кодом или политикой, а не upstream linter configuration.

---

# 13. Integration tests

## 13.1. Общий принцип

Integration tests с PostgreSQL обязательны.

Но инфраструктура integration tests не должна загрязнять `go.mod` core library.

## 13.2. Предпочтительная структура

```text
/
├── go.mod
├── *.go
├── *_test.go
├── integration/
│   ├── go.mod
│   ├── go.sum
│   ├── *_test.go
│   └── README.md
├── docker-compose.test.yml
└── ...
```

`integration/go.mod` является отдельным Go module.

## 13.3. PostgreSQL provisioning

Предпочтительный вариант:

- CI service container PostgreSQL;
- локально — `docker compose`;
- integration tests получают DSN через environment variable.

Например:

```text
TEST_DATABASE_URL
```

или:

```text
POSTGRES_TEST_DSN
```

Не поднимать Docker из Go-кода без необходимости.

## 13.4. Integration module dependencies

Integration module может зависеть от:

```text
github.com/jackc/pgx/v5
```

Не использовать `scany`, если обычный `rows.Scan`/`QueryRow.Scan` достаточен.

Не использовать `testdock` / `dockertest`, если PostgreSQL может быть поднят CI service/container.

---

# 14. Regression test suite: placeholders

Это критическая часть задачи.

Не ограничиваться переносом upstream tests.

Создать отдельный набор regression tests для сложных запросов.

## 14.1. Basic Dollar placeholders

Проверить:

```text
? → $1
?, ? → $1, $2
```

Порядок `args` должен точно соответствовать SQL.

## 14.2. Nested SELECT in WHERE

Проверить subquery с собственными args + args parent query.

## 14.3. Nested SELECT in Eq

Пример:

```go
Eq{
    "account_id": Select("id").
        From("accounts").
        Where("tenant_id = ?", 42),
}
```

## 14.4. Nested SELECT in JOIN

Обязательно regression test для:

```go
JoinClause(...)
```

и/или supported join-subquery API.

Это исторически проблемная область оригинального Squirrel.

## 14.5. FromSelect

Проверить placeholders:

```text
outer prefix args
FROM (subquery with args)
outer WHERE args
```

## 14.6. Column subquery

Проверить:

```sql
SELECT
    ...,
    (SELECT ... WHERE x = ?) AS value
FROM ...
WHERE y = ?
```

## 14.7. UPDATE SET subquery

Проверить:

```sql
UPDATE ...
SET value = (SELECT ... WHERE ...)
WHERE ...
```

## 14.8. And / Or nesting

Минимум 3 уровня вложенности:

```text
AND
 ├── Expr
 └── OR
      ├── nested SELECT
      └── AND
```

## 14.9. CTE placeholders

Проверить args:

```text
CTE #1
CTE #2
final SELECT
```

Ожидание:

```text
$1...$N
```

без reset numbering между CTE.

## 14.10. Recursive CTE

Добавить тест с placeholders в:

- anchor query;
- recursive query;
- final query.

## 14.11. DML CTE

Проверить:

```text
SELECT candidate
UPDATE ... FROM candidate
RETURNING
SELECT FROM updated
```

с args во всех уровнях.

## 14.12. Prefix / Suffix

Проверить placeholders в:

- `Prefix`;
- body;
- `Suffix`.

## 14.13. PostgreSQL JSON operators

Критически проверить escaped question mark syntax Squirrel:

```sql
meta->'format' ??| array[?, ?]
```

после Dollar replacement должен давать корректный PostgreSQL JSON operator и корректные `$N`.

Добавить аналогичные тесты для:

```text
?
?|
?&
```

в тех формах, которые поддерживаются escape convention библиотеки.

## 14.14. Repeated builder reuse

Squirrel builders должны вести себя immutable-style.

Проверить:

```go
base := Select(...).Where(...)
a := base.Where(...)
b := base.Where(...)
```

`a` не должен мутировать `b` или `base`.

## 14.15. Custom Sqlizer

Добавить тест пользовательского типа, реализующего только:

```go
ToSql()
```

Проверить и задокументировать контракт nested custom `Sqlizer`.

Особенно важно определить:

- ожидается ли от custom `Sqlizer` возврат `?`;
- что происходит, если custom Sqlizer уже возвращает `$1`.

Не пытаться магически renumber arbitrary preformatted `$N`.

Рекомендуемый контракт:

> nested custom Sqlizer должен возвращать raw SQL с `?`, если должен участвовать в placeholder formatting родительского builder.

---

# 15. Regression tests PostgreSQL execution

SQL string comparison недостаточен.

Часть запросов должна реально выполняться в PostgreSQL.

Минимальный набор integration scenarios:

1. simple SELECT;
2. nested SELECT;
3. correlated EXISTS;
4. CTE;
5. recursive CTE;
6. `UPDATE ... FROM`;
7. DML CTE;
8. `FOR UPDATE SKIP LOCKED`;
9. `INSERT ... RETURNING`;
10. `UPDATE ... RETURNING`;
11. `DELETE ... RETURNING`;
12. `IN`/`NOT IN`;
13. `ANY`/`ALL`, если fork сохраняет это поведение;
14. JSON operators;
15. CASE;
16. COALESCE.

Цель integration tests:

не просто проверить синтаксическую строку, а доказать, что generated SQL принимается PostgreSQL и bind args работают через pgx.

---

# 16. Fuzz testing

Добавить Go fuzz tests для placeholder replacement/composition.

Минимальная цель:

- отсутствие panic;
- отсутствие invalid placeholder numbering;
- количество placeholders соответствует количеству args для корректных генерируемых cases;
- escaped `??` не расходует bind argument;
- последовательность Dollar placeholders не содержит пропусков.

Fuzz tests не должны пытаться полноценно парсить arbitrary SQL.

Фокус — именно на механике placeholder transformation.

---

# 17. Error handling

Проверить и унифицировать ошибки:

- SELECT без columns;
- INSERT без table;
- INSERT без values/select;
- invalid CTE;
- CTE без final statement;
- unsupported expression type;
- invalid nested Sqlizer;
- placeholder replacement errors.

Не вводить собственную сложную hierarchy ошибок без необходимости.

Ошибки должны:

- быть понятными;
- содержать контекст;
- не включать sensitive args.

---

# 18. SQL injection boundaries

README должен явно разделять:

## Values

Values должны передаваться через placeholders:

```go
Where("user_id = ?", userID)
```

## Identifiers / raw SQL

API вида:

```go
From(userInput)
OrderBy(userInput)
Column(userInput)
Expr(userInput)
```

не должны позиционироваться как безопасные для untrusted input.

Не добавлять автоматический identifier sanitizer в рамках этой задачи.

Документировать boundary.

---

# 19. README

Переписать README под новый scope.

README должен содержать:

1. что это maintained internal/minimal fork Squirrel;
2. attribution:
   - Masterminds/squirrel;
   - n-r-w/squirrel;
3. rationale fork;
4. pure SQL builder scope;
5. supported Go version;
6. dependency policy;
7. PostgreSQL focus;
8. nested query examples;
9. CTE example;
10. DML CTE example;
11. `UPDATE ... FROM`;
12. placeholder rules;
13. custom Sqlizer contract;
14. migration notes с Masterminds;
15. список намеренно удалённого API.

Не копировать README upstream целиком без необходимости.

---

# 20. License / attribution

Исходные проекты используют MIT license.

Обязательно:

- сохранить `LICENSE`;
- сохранить требуемые copyright notices;
- не удалять attribution исходных авторов;
- в README явно написать, что проект основан на:
  - `Masterminds/squirrel`;
  - `n-r-w/squirrel`.

Если при fork изменяется copyright header / notice, сделать это корректно без удаления предыдущих правообладателей.

---

# 21. CI

Минимальный pipeline:

```text
format
↓
go vet
↓
lint
↓
unit tests
↓
race tests
↓
fuzz smoke
↓
integration tests PostgreSQL
↓
dependency verification
```

## 21.1. Commands

Минимум:

```bash
gofmt -w / check
go vet ./...
go test ./...
go test -race ./...
go mod tidy
go mod verify
```

Lint использовать только если linter уже является корпоративным стандартом.

Не добавлять десятки инструментов в module dependencies.

CLI tools должны устанавливаться CI отдельно и не попадать в `go.mod` library.

---

# 22. Dependency verification в CI

Добавить автоматическую проверку, не позволяющую dependency graph снова разрастись.

Например отдельный script:

```bash
go list -m all
```

и policy check.

Core module должен падать в CI, если появляются запрещённые категории зависимостей.

Как минимум запретить:

```text
docker
dockertest
testdock
scany
pgx
mysql
mongo
goose
migrate
```

в root module graph.

Важно:

`pgx` разрешён в `integration/` module, но не в root.

---

# 23. API surface audit

Перед удалением кода сгенерировать inventory exported API:

```bash
go doc
```

или небольшим AST-based script.

Сохранить результаты для:

```text
Masterminds v1.5.4
n-r-w v1.6.0
new fork
```

Составить таблицу:

| Symbol | Masterminds | n-r-w | Fork | Decision |
|---|---:|---:|---:|---|
| Select | yes | yes | yes | keep |
| RunWith | yes | no | no | intentionally removed |
| CTE API | no | yes | yes | keep |
| Search | no | yes | no | application helper |
| Paginator | no | yes | no | application helper |

Не обязательно вручную перечислять каждый symbol в README, но итоговый технический отчёт должен содержать полный diff exported API.

---

# 24. Migration compatibility test project

Создать небольшой compile-only test fixture, имитирующий типичное использование Masterminds.

Проверить, что после изменения import path продолжают компилироваться основные patterns:

```go
sq.Select(...)
sq.Insert(...)
sq.Update(...)
sq.Delete(...)
sq.Eq{}
sq.And{}
sq.Or{}
sq.Expr(...)
sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
```

Отдельно перечислить intentional incompatibilities:

- DB execution API;
- StmtCache;
- удалённые/изменённые Case semantics, если они всё-таки останутся;
- другие breaking differences.

---

# 25. Не переносить изменения механически

Нельзя просто:

```text
fork n-r-w → delete go.mod lines → done
```

Агент должен разобраться, какие exported features создают лишние зависимости и какой слой ответственности они представляют.

Особенно внимательно проверить:

- `select.go`;
- `expr.go`;
- `cte.go`;
- `part.go`;
- `placeholder.go`;
- `statement.go`;
- `case.go`;
- `update.go`.

---

# 26. Рекомендуемая последовательность реализации

## Phase 1 — Baseline capture

1. Checkout `n-r-w/squirrel v1.6.0`.
2. Зафиксировать commit baseline.
3. Запустить все upstream unit tests.
4. Зафиксировать:
   - exported API;
   - `go.mod`;
   - `go mod graph`;
   - current coverage.

## Phase 2 — Dependency cleanup

1. Удалить/вынести integration test infrastructure.
2. Удалить `pgx`, `scany`, `testdock` из root module.
3. Удалить `testify` из root, если возможно.
4. Удалить `x/exp`.
5. `go mod tidy`.
6. Проверить dependency graph.

## Phase 3 — Scope cleanup

Удалить:

- Search;
- Paginator;
- Paginate*;
- OrderByCond;
- EqNotEmpty;
- application-level alias helpers.

После каждого логического удаления прогонять unit tests.

## Phase 4 — Compatibility normalization

1. Сравнить core API с Masterminds v1.5.4.
2. Проверить Case behavior.
3. Восстановить совместимое поведение там, где это разумно.
4. Зафиксировать intentional breaking changes.

## Phase 5 — Nested-query hardening

1. Сохранить `rawSqlizer`.
2. Проверить все builders на `toSqlRaw`.
3. Убедиться, что nested composition использует `nestedToSql`.
4. Добавить regression suite.
5. Добавить fuzz tests.

## Phase 6 — PostgreSQL extensions

Проверить и стабилизировать:

- CTE;
- recursive CTE;
- DML CTE;
- `UPDATE ... FROM`;
- EXISTS;
- comparisons;
- COALESCE;
- IN/NOT IN.

## Phase 7 — Integration module

1. Создать `integration/go.mod`.
2. Добавить pgx.
3. Создать schema/test fixtures.
4. Добавить real PostgreSQL tests.
5. Настроить CI PostgreSQL service.

## Phase 8 — Documentation

1. README.
2. Migration guide.
3. API diff.
4. Dependency rationale.
5. License attribution.

## Phase 9 — Final validation

Полный:

```bash
go test ./...
go test -race ./...
go vet ./...
go mod tidy
go mod verify
```

Плюс integration tests и dependency policy check.

---

# 27. Definition of Done

Задача считается выполненной, если выполнены ВСЕ условия.

## Core architecture

- [ ] библиотека только строит SQL;
- [ ] нет DB execution/scanning API;
- [ ] nested query placeholders обрабатываются до единственной финальной numbering phase;
- [ ] `Dollar` placeholders корректны в complex nested queries.

## Dependencies

- [ ] root `go.mod` не содержит pgx;
- [ ] root `go.mod` не содержит scany;
- [ ] root `go.mod` не содержит testdock;
- [ ] root module graph не содержит Docker ecosystem;
- [ ] root не зависит от `golang.org/x/exp`;
- [ ] runtime dependency фактически ограничена `lann/builder` и его минимальными transitive deps.

## Features

- [ ] basic Squirrel builders работают;
- [ ] CTE работает;
- [ ] recursive CTE работает;
- [ ] DML CTE работает;
- [ ] `UPDATE ... FROM` работает;
- [ ] nested subqueries работают;
- [ ] EXISTS работает;
- [ ] COALESCE работает;
- [ ] PostgreSQL Dollar placeholders работают.

## Removed scope

- [ ] Search отсутствует;
- [ ] Paginator отсутствует;
- [ ] Paginate* отсутствует;
- [ ] OrderByCond отсутствует;
- [ ] EqNotEmpty отсутствует.

## Tests

- [ ] upstream relevant tests проходят;
- [ ] regression placeholder suite добавлен;
- [ ] integration tests реально выполняются в PostgreSQL;
- [ ] race test проходит;
- [ ] fuzz smoke test проходит;
- [ ] migration compile fixture проходит.

## Documentation

- [ ] README обновлён;
- [ ] migration guide есть;
- [ ] intentional breaking changes перечислены;
- [ ] attribution сохранён;
- [ ] dependency policy описан.

---

# 28. Ожидаемые артефакты от агента

В результате задачи предоставить:

1. готовый repository fork;
2. очищенный `go.mod`;
3. unit tests;
4. fuzz tests;
5. PostgreSQL integration module/tests;
6. CI configuration;
7. README;
8. `MIGRATION.md`;
9. `UPSTREAM.md` или аналогичный файл с информацией:
   - исходный Masterminds version;
   - n-r-w baseline;
   - какие изменения сохранены;
   - какие изменения удалены;
10. короткий итоговый отчёт.

---

# 29. Формат итогового отчёта агента

В конце работы агент должен вывести:

```markdown
## Result

### Baseline
- n-r-w/squirrel: v1.6.0 / <commit>

### Core dependencies before
...

### Core dependencies after
...

### Preserved upstream features
...

### Removed features
...

### Compatibility differences vs Masterminds v1.5.4
...

### Regression bugs covered
...

### PostgreSQL integration scenarios
...

### Tests
- unit:
- race:
- fuzz:
- integration:

### Remaining risks
...

### Recommended follow-up work
...
```

---

# 30. Out of scope

В рамках этой задачи НЕ делать:

- ORM;
- struct mapper;
- repository layer;
- query executor;
- transaction helpers;
- connection pooling;
- pgx wrapper;
- automatic scanning;
- migration engine;
- generic database abstraction;
- query parser;
- SQL AST rewrite;
- replacement `lann/builder`;
- broad redesign всего Squirrel API;
- application pagination;
- search DSL;
- automatic identifier escaping based on user input.

---

# 31. Follow-up tasks, которые можно рассмотреть отдельно

После стабилизации первой версии можно отдельно оценить:

## 31.1. Удаление lann/builder

Переписать builders на обычные typed structs/copy-on-write API.

Делать только после того, как regression suite гарантирует совместимость.

## 31.2. PostgreSQL-specific package

Если PG helpers начнут расти, вместо загрязнения root API создать:

```text
squirrel
squirrel/postgres
```

Например туда потенциально могут попасть:

- array ANY/ALL;
- JSONB helpers;
- DISTINCT ON;
- ON CONFLICT helpers;
- locking clauses;
- PostgreSQL-specific RETURNING API.

Не делать это заранее без реального use case.

## 31.3. Upstream synchronization process

Определить периодический процесс просмотра:

- Masterminds PR/issues;
- n-r-w releases/commits;
- security advisories.

Не делать автоматический merge upstream.

Каждое upstream изменение должно проходить через:

```text
relevance
→ dependency impact
→ compatibility impact
→ regression tests
→ cherry-pick/manual port
```

---

# 32. Источники и baseline notes

Актуальность исходных данных: август 2026.

## Masterminds/squirrel

Repository:

https://github.com/Masterminds/squirrel

Releases:

https://github.com/Masterminds/squirrel/releases

Ключевые исторические releases:

- `v1.5.1` — исправления Select subquery + DollarPlaceholder;
- `v1.5.2` — исправления placeholder generation для And/Or;
- `v1.5.3` — исправления JoinClause/subquery dollar placeholders;
- `v1.5.4` — последний опубликованный release оригинального проекта.

## n-r-w/squirrel

Repository:

https://github.com/n-r-w/squirrel

Releases:

https://github.com/n-r-w/squirrel/releases

Baseline:

```text
v1.6.0
commit 4c87dbb
```

Ключевые изменения:

- `v1.4.3` — fix parameter numbering for nested queries;
- `v1.5.0` — добавлены integration tests, после чего значительно вырос `go.mod`;
- `v1.5.1` — refactor integration tests;
- `v1.6.0` — Go 1.26 / DML CTE enhancements.

Текущий код nested composition:

https://github.com/n-r-w/squirrel/blob/master/part.go

Текущий `go.mod`:

https://github.com/n-r-w/squirrel/blob/master/go.mod

README с extensions:

https://github.com/n-r-w/squirrel/blob/master/README.md

---

# 33. Краткая архитектурная формулировка для агента

Если требуется держать одну мысль во время реализации, использовать следующую:

> Построй маленький maintained fork Squirrel: сохрани знакомый DSL Masterminds, возьми из n-r-w исправленную композицию nested queries и полезные SQL/PostgreSQL primitives, но удали database execution, integration infrastructure dependencies и application-level helpers. Главный критерий качества — корректный SQL и placeholders в сложных запросах при минимальном dependency graph.

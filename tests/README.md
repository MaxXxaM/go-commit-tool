# E2E Tests для gomm

Этот пакет содержит полный набор end-to-end (E2E) тестов для gomm - утилиты управления мультимодульными Go проектами.

## Структура тестов

```
tests/
├── e2e/                          # End-to-end тесты
│   ├── 01_init_test.go          # Тесты инициализации (init, scan, status, tree)
│   ├── 02_clone_test.go         # Тесты клонирования (clone, move, clean)
│   ├── 03_workspace_test.go     # Тесты workspace (local, remote, replace)
│   ├── 04_versioning_test.go    # Тесты версионирования (version check)
│   ├── 05_release_test.go       # Тесты релизов (release)
│   └── 06_full_workflow_test.go # Полные workflow сценарии
├── helpers/
│   └── testutils.go             # Вспомогательные функции для тестов
└── fixtures/                    # Тестовые данные (создаются динамически)
```

## Покрытие тестами

### 1. Инициализация и анализ (01_init_test.go)

Тестирует базовые команды для работы с проектом:

- `TestInitCommand` - инициализация проекта
- `TestInitWithExistingConfig` - повторная инициализация
- `TestScanCommand` - сканирование зависимостей
- `TestScanWithDepth` - сканирование с ограничением глубины
- `TestStatusCommand` - проверка статуса проекта
- `TestStatusVerbose` - детальный статус
- `TestTreeCommand` - визуализация дерева зависимостей
- `TestTreeWithDepthLimit` - дерево с ограничением глубины
- `TestTreeLocalOnly` - отображение только локальных модулей
- `TestConfigWithSharedDir` - настройка shared_dir
- `TestUninitializedProject` - работа в неинициализированном проекте

### 2. Клонирование и размещение (02_clone_test.go)

Тестирует управление локальными копиями зависимостей:

- `TestCloneSingleDependency` - клонирование одной зависимости
- `TestCloneWithLocalMode` - клонирование в local режиме
- `TestCloneWithSharedMode` - клонирование в shared режиме
- `TestCloneAll` - клонирование всех зависимостей
- `TestCloneAlreadyCloned` - повторное клонирование
- `TestMoveLocalToShared` - перемещение из local в shared
- `TestMoveSharedToLocal` - перемещение из shared в local
- `TestCleanSingleDependency` - удаление зависимости
- `TestCleanWithUncommittedChanges` - удаление с несохраненными изменениями
- `TestCleanAll` - удаление всех зависимостей
- `TestCloneNonExistentDependency` - клонирование несуществующей зависимости
- `TestPlacementRules` - автоматическое определение placement

### 3. Workspace управление (03_workspace_test.go)

Тестирует работу с go.work и переключение режимов:

- `TestLocalCommand` - переключение на локальную разработку
- `TestLocalAllDependencies` - переключение всех на local
- `TestLocalWithPrefix` - переключение по префиксу
- `TestRemoteCommand` - переключение на remote
- `TestRemoteAllDependencies` - переключение всех на remote
- `TestMixedMode` - смешанный режим работы
- `TestReplaceAdd` - добавление custom replace
- `TestReplaceAddLocalPath` - replace на локальный путь
- `TestReplaceList` - отображение replace директив
- `TestReplaceRemove` - удаление replace
- `TestWorkspacePreservesCustomReplaces` - сохранение custom replace
- `TestLocalWithoutClone` - попытка local без clone
- `TestWorkspaceWithNestedDependencies` - workspace с вложенными зависимостями
- `TestStatusShowsWorkspaceInfo` - отображение workspace в status

### 4. Версионирование (04_versioning_test.go)

Тестирует анализ версий и определение следующей версии:

- `TestVersionCheckCurrentVersion` - отображение текущей версии
- `TestVersionCheckNoChanges` - версия без изменений
- `TestVersionCheckWithMinorChanges` - определение minor версии
- `TestVersionCheckWithPatchChanges` - определение patch версии
- `TestVersionCheckForSpecificModule` - version check для модуля
- `TestVersionCheckUntaggedModule` - модуль без тегов
- `TestVersionCheckShowsChangesSummary` - сводка изменений
- `TestVersionCheckVerbose` - verbose режим
- `TestVersionCheckWithMultipleVersionTags` - модуль с несколькими тегами
- `TestVersionCheckInProjectRoot` - version check в корне проекта
- `TestVersionCheckDetectsBreakingChanges` - определение breaking changes

### 5. Релизы (05_release_test.go)

Тестирует создание релизов и обновление версий:

- `TestReleaseAuto` - автоматический релиз
- `TestReleaseBumpPatch` - релиз с patch версией
- `TestReleaseBumpMinor` - релиз с minor версией
- `TestReleaseBumpMajor` - релиз с major версией
- `TestReleaseWithSpecificVersion` - релиз с конкретной версией
- `TestReleaseWithoutChanges` - релиз без изменений (ошибка)
- `TestReleaseUpdatesGoMod` - обновление go.mod после релиза
- `TestReleaseCascade` - каскадный релиз
- `TestReleaseNoTests` - релиз без запуска тестов
- `TestReleaseFailsOnFailedTests` - провал при ошибке тестов
- `TestReleaseForSpecificModule` - релиз конкретного модуля
- `TestReleaseCreatesProperTag` - проверка формата тега
- `TestReleaseWithUncommittedChanges` - релиз с несохраненными изменениями
- `TestReleaseFromProjectRoot` - релиз самого проекта

### 6. Полные workflow (06_full_workflow_test.go)

Тестирует реальные сценарии использования от начала до конца:

- `TestCompleteWorkflow` - полный workflow от инициализации до релиза
- `TestMultiServiceWorkflow` - работа с несколькими сервисами
- `TestHotfixWorkflow` - workflow быстрого исправления
- `TestFeatureBranchWorkflow` - разработка новой фичи
- `TestMigrationWorkflow` - перемещение модуля между режимами
- `TestCleanupWorkflow` - очистка после завершения работы
- `TestReplaceWorkflow` - работа с custom replace директивами
- `TestPartialWorkspaceWorkflow` - частичное использование workspace

## Запуск тестов

### Предварительные требования

1. Go 1.21+
2. Git
3. Собранный бинарник gomm

### Запуск всех тестов

```bash
# Запуск всех E2E тестов
make test-e2e

# Запуск всех тестов (unit + E2E)
make test-all
```

### Запуск отдельных групп тестов

```bash
# Тесты инициализации
make test-e2e-init

# Тесты клонирования
make test-e2e-clone

# Тесты workspace
make test-e2e-workspace

# Тесты версионирования
make test-e2e-versioning

# Тесты релизов
make test-e2e-release

# Полные workflow тесты
make test-e2e-workflow
```

### Запуск конкретного теста

```bash
# Используя go test напрямую
GOMM_BINARY=./bin/gomm go test -v -timeout 10m ./tests/e2e/ -run TestCompleteWorkflow

# Или собрать и запустить
make build
GOMM_BINARY=$(pwd)/bin/gomm go test -v ./tests/e2e/ -run TestCloneSingleDependency
```

## Как работают тесты

### Тестовое окружение

Каждый тест создает изолированное окружение:

1. **Временная директория** - все тесты работают в отдельной temp директории
2. **Тестовые репозитории** - динамически создаются git репозитории для тестирования
3. **Проектная структура** - симулируется реальная структура проекта с зависимостями
4. **Автоматическая очистка** - после завершения теста все временные файлы удаляются

### Структура тестового окружения

```
temp-dir/
├── project/              # Тестовый проект
│   ├── go.mod
│   ├── main.go
│   ├── .gomm.yaml       # Конфиг gomm
│   └── .gomm/           # Метаданные
├── shared/              # Shared зависимости
│   └── common-lib/
└── test-repos/          # Исходные репозитории зависимостей
    ├── dep-a/
    └── dep-b/
```

### Вспомогательные функции

`tests/helpers/testutils.go` предоставляет набор утилит:

- `NewTestEnv(t)` - создание тестового окружения
- `RunGomm()` - выполнение команд gomm
- `CreateTestModule()` - создание тестового модуля
- `CreateTestProject()` - создание тестового проекта
- `AssertFileExists()` - проверка существования файла
- `AssertContains()` - проверка содержимого вывода
- И многие другие...

## Типичные сценарии тестирования

### 1. Базовый workflow разработчика

```go
func TestCompleteWorkflow(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    // 1. Создание зависимостей
    env.CreateTestModule("github.com/test/lib")
    env.CreateTestProject("github.com/test/service", "github.com/test/lib")

    // 2. Инициализация
    env.RunGommExpectSuccess("init")
    env.RunGommExpectSuccess("scan")

    // 3. Клонирование
    env.RunGommExpectSuccess("clone", "github.com/test/lib")

    // 4. Локальная разработка
    env.RunGommExpectSuccess("local", "github.com/test/lib")

    // 5. Внесение изменений
    // ...

    // 6. Релиз
    env.RunGommExpectSuccess("release", "--bump", "minor")

    // 7. Возврат к remote
    env.RunGommExpectSuccess("remote", "--all")
}
```

### 2. Тестирование ошибочных ситуаций

```go
func TestCloneNonExistentDependency(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    env.CreateTestProject("github.com/test/project")
    env.RunGommExpectSuccess("init")

    // Ожидаем ошибку
    output := env.RunGommExpectError("clone", "github.com/nonexistent/module")
    env.AssertContains(output, "not found")
}
```

## Отладка тестов

### Включение verbose режима

```bash
go test -v ./tests/e2e/ -run TestName
```

### Сохранение временных файлов

Измените `t.TempDir()` в `testutils.go` на создание директории в `/tmp` для инспекции:

```go
tempDir := "/tmp/gomm-test-" + t.Name()
os.MkdirAll(tempDir, 0755)
```

### Логирование команд

Используйте `t.Logf()` для отладочного вывода:

```go
output := env.RunGomm("status")
t.Logf("Status output:\n%s", output)
```

## CI/CD интеграция

Тесты можно интегрировать в CI/CD pipeline:

```yaml
# GitHub Actions example
- name: Run E2E tests
  run: |
    make build
    make test-e2e
  timeout-minutes: 30
```

## Расширение тестов

### Добавление нового теста

1. Определите категорию теста (init, clone, workspace, версионирование, релиз, workflow)
2. Добавьте тест в соответствующий файл
3. Используйте helpers из `testutils.go`
4. Следуйте паттерну: Setup -> Action -> Assert -> Cleanup

Пример:

```go
func TestNewFeature(t *testing.T) {
    env := helpers.NewTestEnv(t)
    defer env.Cleanup()

    // Setup
    env.CreateTestProject("github.com/test/project")
    env.RunGommExpectSuccess("init")

    // Action
    output := env.RunGommExpectSuccess("new-command", "--flag")

    // Assert
    env.AssertContains(output, "expected text")
    env.AssertFileExists("/path/to/file")
}
```

### Добавление новых helper функций

Если вам нужна новая вспомогательная функция, добавьте её в `testutils.go`:

```go
func (env *TestEnv) YourNewHelper() {
    env.T.Helper()
    // Implementation
}
```

## Покрытие workflow сценариев

Тесты покрывают следующие workflow:

✅ Инициализация проекта
✅ Сканирование и анализ зависимостей
✅ Клонирование зависимостей (local/shared)
✅ Перемещение между local и shared
✅ Переключение на локальную разработку (go.work)
✅ Смешанный режим (partial workspace)
✅ Custom replace директивы
✅ Версионирование и определение типа версии
✅ Создание релизов с различными опциями
✅ Каскадный релиз зависимостей
✅ Hotfix workflow
✅ Feature development workflow
✅ Cleanup и удаление зависимостей

## Примечания

- Тесты используют реальные git операции (init, commit, tag)
- Каждый тест полностью изолирован
- Тесты не требуют доступа к интернету (все репозитории локальные)
- Время выполнения всех E2E тестов: ~10-30 минут
- Для ускорения можно запускать тесты параллельно с `-parallel` флагом

## Известные ограничения

1. Некоторые тесты могут пропускаться (skip), если функционал еще не реализован
2. Тесты версионирования используют упрощенную логику определения типа версии
3. Каскадный релиз тестируется в базовом варианте

## Поддержка

Если у вас возникли проблемы с тестами:

1. Убедитесь, что gomm собран: `make build`
2. Проверьте версию Go: `go version` (требуется 1.21+)
3. Запустите отдельный тест с verbose: `go test -v ./tests/e2e/ -run TestName`
4. Проверьте логи в выводе теста

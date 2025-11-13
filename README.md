# gomm - Go Multi-Module Manager

Утилита для управления разработкой в мультимодульных Go проектах со сложными зависимостями.

## Возможности

- 🔍 Анализ и визуализация дерева зависимостей
- 🔄 Переключение между локальной разработкой (go.work) и удаленными версиями
- 📦 Управление локальными копиями зависимостей (local/shared)
- 🏷️ Автоматическое версионирование с анализом изменений
- 🚀 Автоматизация релизов и обновления зависимостей
- 🔀 Миграция модулей между репозиториями

## Установка

### Из исходников

```bash
git clone https://github.com/MaxXxaM/gomm.git
cd gomm
make install
```

Или просто:

```bash
go install github.com/MaxXxaM/gomm/cmd/gomm@latest
```

## Быстрый старт

```bash
# 1. Инициализация в проекте
cd your-go-project
gomm init

# 2. Сканирование зависимостей
gomm scan

# 3. Визуализация дерева
gomm tree

# 4. Проверка статуса
gomm status
```

## Команды

### Инициализация и конфигурация

```bash
# Инициализация проекта
gomm init

# Сканирование зависимостей
gomm scan [--depth N]

# Показать статус проекта
gomm status [-v]

# Визуализация дерева зависимостей
gomm tree [--depth N] [--local-only]
```

### Управление зависимостями

```bash
# Клонирование зависимостей
gomm clone <module-path> [--mode local|shared|auto]

# Клонирование всех зависимостей
gomm clone --all

# Перемещение модуля между local/shared
gomm move <module-path> --to local|shared

# Удаление локальной копии
gomm clean <module-path>
gomm clean <module-path> --force
gomm clean --all

# Переключение на локальную разработку (go.work)
gomm local <module-path>...
gomm local --all
gomm local --prefix github.com/myorg

# Переключение на удаленные версии (go.mod)
gomm remote <module-path>...
gomm remote --all
gomm remote --prefix github.com/myorg

# Управление replace директивами
gomm replace add <module> <replacement>
gomm replace remove <module>
gomm replace list
```

### Версионирование и релизы

```bash
# Проверка текущей версии и предложение следующей
gomm version check
gomm version check github.com/user/repo

# Создание релиза с автоопределением версии
gomm release
gomm release --auto

# Создание релиза с указанием типа версии
gomm release --bump patch
gomm release --bump minor
gomm release --bump major

# Создание релиза с конкретной версией
gomm release --version v1.2.3

# Создание релиза для конкретного модуля
gomm release github.com/user/repo --bump minor

# Каскадный релиз (обновление зависимых модулей)
gomm release --cascade

# Релиз без запуска тестов
gomm release --no-tests
```

## Конфигурация

При выполнении `gomm init` создается файл `.gomm.yaml`:

```yaml
# Директория для shared зависимостей
shared_dir: ~/shared

# Стратегия размещения по умолчанию
default_placement: auto  # auto, local, shared

# Правила автоматического определения placement
placement_rules:
  - pattern: "github.com/myorg/common-*"
    mode: shared
  - pattern: "github.com/myorg/service-*"
    mode: local

# Настройки версионирования
versioning:
  auto_detect: true

# Настройки релиза
release:
  auto_push: true
  create_changelog: false
  run_tests: true

# Исключения (не обрабатывать эти модули)
exclude:
  - "golang.org/x/*"
  - "github.com/golang/*"
  - "std"
```

## Workflow

### Базовый workflow разработки

```bash
# 1. Инициализация и сканирование
gomm init
gomm scan

# 2. Клонирование нужных зависимостей
gomm clone github.com/myorg/lib-a
gomm clone github.com/myorg/lib-b --mode shared

# 3. Переключение на локальную разработку
gomm local github.com/myorg/lib-a github.com/myorg/lib-b

# 4. Разработка...
# Внесите изменения в lib-a и lib-b

# 5. Проверка статуса
gomm status

# 6. Создание релизов (автоматическое определение версий)
gomm release github.com/myorg/lib-a --auto
gomm release github.com/myorg/lib-b --auto

# 7. Переключение обратно на удаленные версии
gomm remote --all
```

### Работа с workspace

gomm автоматически управляет `go.work` файлом:

- `gomm local` - добавляет модули в go.work (локальная разработка)
- `gomm remote` - удаляет из go.work (удаленные версии)
- Поддержка частичного переключения (смешанный режим)

#### Смешанный режим

Вы можете работать с некоторыми модулями локально, а другие использовать из go.mod:

```bash
# Переключить только один модуль на локальную разработку
gomm local github.com/myorg/lib-a

# Остальные модули останутся в remote режиме
gomm status
# В режиме workspace: 1
# В режиме remote: 2
#   (смешанный режим)
```

## Структура проекта

```
.
├── cmd/
│   └── gomm/              # Main приложение
├── internal/
│   ├── analyzer/          # Анализ go.mod и построение дерева
│   ├── cloner/            # Клонирование репозиториев
│   ├── workspace/         # Управление go.work
│   ├── versioning/        # Версионирование
│   ├── config/            # Конфигурация
│   ├── git/               # Git операции
│   └── cli/               # CLI команды
├── pkg/
│   ├── types/             # Общие типы
│   └── logger/            # Логирование
├── .gomm.yaml             # Конфигурация (создается при init)
├── .gomm/                 # Метаданные проекта
│   └── metadata.json
└── README.md
```

## Разработка

### Требования

- Go 1.21+
- Git

### Сборка

```bash
# Сборка
make build

# Установка
make install

# Тесты
make test

# Форматирование
make fmt

# Линтер
make lint

# Все команды
make help
```

### Roadmap

**Итерация 1** (завершена):
- ✅ Базовая инфраструктура
- ✅ Анализ go.mod и построение дерева зависимостей
- ✅ Команды: init, scan, status, tree
- ✅ Конфигурация

**Итерация 2** (завершена):
- ✅ Git операции (клонирование, проверка состояния)
- ✅ Клонирование и управление локальными копиями
- ✅ Команды: clone, move, clean
- ✅ Управление placement (local/shared)
- ✅ Метаданные модулей

**Итерация 3** (завершена):
- ✅ WorkspaceManager для управления go.work
- ✅ Команды local/remote для переключения режимов
- ✅ Custom replace директивы
- ✅ Смешанный режим (частичное переключение)
- ✅ Обновленная команда status с информацией о workspace

**Итерации 5-6** (завершены):
- ✅ Version Analyzer для определения версий
- ✅ Анализ изменений (базовый)
- ✅ Автоматическое определение типа версии (major/minor/patch)
- ✅ Команда version check
- ✅ ReleaseManager для создания релизов
- ✅ Команда release с поддержкой --auto, --bump, --version
- ✅ Создание и push git тегов
- ✅ Обновление зависимостей в go.mod
- ✅ Базовая поддержка каскадного релиза

**Будущие итерации:**

**Итерация 4** (отложена):
- Аутентификация (SSH, токены)
- Работа с приватными репозиториями

**Итерация 7**:
- Миграция репозиториев
- Обновление namespace

Полный план разработки см. в [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md)

## Документация

- [Техническое задание](TECHNICAL_SPEC.md) - полное описание функциональности
- [План разработки](DEVELOPMENT_PLAN.md) - детальный план по итерациям

## Примеры использования

### Пример 1: Работа с зависимостью

```bash
# Клонировать зависимость локально
gomm clone github.com/myorg/auth-lib

# Переключиться на локальную версию
gomm local github.com/myorg/auth-lib

# Внести изменения в auth-lib...

# Создать релиз
gomm release github.com/myorg/auth-lib --auto

# Вернуться на удаленную версию
gomm remote github.com/myorg/auth-lib
```

### Пример 2: Клонирование и управление зависимостями

```bash
# Клонировать конкретную зависимость (auto режим)
gomm clone github.com/myorg/common-lib

# Клонировать в shared директорию
gomm clone github.com/myorg/shared-utils --mode shared

# Клонировать все зависимости проекта
gomm clone --all

# Переместить модуль из local в shared
gomm move github.com/myorg/common-lib --to shared

# Проверить статус
gomm status

# Удалить локальную копию
gomm clean github.com/myorg/temp-lib

# Удалить все локальные копии
gomm clean --all --force
```

### Пример 3: Работа с workspace (go.work)

```bash
# Клонировать зависимости
gomm clone github.com/myorg/lib-a github.com/myorg/lib-b

# Переключить в workspace режим
gomm local github.com/myorg/lib-a github.com/myorg/lib-b

# Проверить что go.work создан
ls go.work

# Разработка с локальными версиями...
# Go автоматически использует локальные версии модулей

# Переключить только один модуль обратно
gomm remote github.com/myorg/lib-a

# Статус покажет смешанный режим
gomm status

# Переключить все обратно на remote
gomm remote --all

# go.work будет удален
```

### Пример 4: Custom replace директивы

```bash
# Добавить replace для модуля
gomm replace add example.com/old-pkg example.com/new-pkg

# Добавить replace на локальный путь
gomm replace add example.com/module ../local-fork

# Показать все replace
gomm replace list

# Удалить replace
gomm replace remove example.com/old-pkg
```

### Пример 5: Визуализация дерева

```bash
# Полное дерево
gomm tree

# Только локальные модули
gomm tree --local-only

# Ограничить глубину
gomm tree --depth 3
```

### Пример 6: Версионирование и релизы

```bash
# Проверить текущую версию и предложение
gomm version check

# Вывод:
# Текущая версия: v0.3.0
# Предлагаемая версия: v0.4.0 (minor)

# Создать релиз с автоопределением версии
gomm release

# Создать релиз для конкретного модуля
cd ../my-library
gomm release --auto

# Или указать конкретный тип bump
gomm release --bump patch   # v0.3.0 -> v0.3.1
gomm release --bump minor   # v0.3.0 -> v0.4.0
gomm release --bump major   # v0.3.0 -> v1.0.0

# Создать релиз с конкретной версией
gomm release --version v2.0.0

# Пропустить тесты (не рекомендуется)
gomm release --no-tests

# После релиза:
# - Создан тег v0.4.0
# - Тег запушен в origin (если auto_push: true)
# - Обновлены метаданные
```

## Вклад в проект

Приветствуются pull requests и issues!

1. Fork проекта
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменений (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## Лицензия

MIT

## Автор

MaxXxaM

## Текущий статус

**Версия:** 0.4.0 (Iterations 5-6)

**Что работает:**
- ✅ Инициализация проекта
- ✅ Сканирование и анализ зависимостей
- ✅ Визуализация дерева зависимостей
- ✅ Базовая конфигурация
- ✅ Клонирование зависимостей локально
- ✅ Управление размещением модулей (local/shared)
- ✅ Перемещение модулей между режимами
- ✅ Удаление локальных копий
- ✅ Метаданные и отслеживание состояния
- ✅ Управление go.work файлом
- ✅ Переключение между workspace и remote режимами
- ✅ Смешанный режим работы
- ✅ Custom replace директивы
- ✅ Анализ версий и автоопределение
- ✅ Создание релизов с тегами
- ✅ Обновление зависимостей в go.mod
- ✅ Запуск тестов перед релизом

**Планируется:**
- Расширенный анализ breaking changes через AST
- Полноценный каскадный релиз
- Миграция репозиториев (Iteration 7)
- Аутентификация для приватных репозиториев

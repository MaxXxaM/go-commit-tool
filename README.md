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

### Управление зависимостями (в разработке)

```bash
# Клонирование зависимостей
gomm clone [--mode local|shared]

# Переключение на локальную разработку
gomm local [module...]

# Переключение на удаленные версии
gomm remote [module...]

# Управление replace директивами
gomm replace add <module> <replacement>
gomm replace remove <module>
gomm replace list
```

### Версионирование и релизы (в разработке)

```bash
# Проверка изменений и предложение версии
gomm version check [module]

# Создание релиза
gomm release [module] [--auto|--bump major|minor|patch|--version v1.2.3]

# Обновление зависимостей
gomm update-deps [--all]
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

- `gomm local` - добавляет модули в go.work
- `gomm remote` - удаляет из go.work
- Поддержка частичного переключения

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

**Итерация 1** (текущая - завершена):
- ✅ Базовая инфраструктура
- ✅ Анализ go.mod и построение дерева зависимостей
- ✅ Команды: init, scan, status, tree
- ✅ Конфигурация

**Итерация 2** (в разработке):
- Клонирование и управление локальными копиями
- Команда clone
- Управление placement (local/shared)

**Итерация 3**:
- Управление go.work
- Команды local/remote
- Custom replace директивы

**Итерация 4**:
- Аутентификация (SSH, токены)
- Работа с приватными репозиториями

**Итерация 5**:
- Анализ изменений
- Автоматическое определение версий

**Итерация 6**:
- Создание релизов
- Каскадное обновление зависимостей

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

### Пример 2: Визуализация дерева

```bash
# Полное дерево
gomm tree

# Только локальные модули
gomm tree --local-only

# Ограничить глубину
gomm tree --depth 3
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

**Версия:** 0.1.0 (Iteration 1)

**Что работает:**
- Инициализация проекта
- Сканирование и анализ зависимостей
- Визуализация дерева зависимостей
- Базовая конфигурация

**В разработке:**
- Клонирование зависимостей (Iteration 2)
- Управление workspace (Iteration 3)
- Версионирование и релизы (Iterations 5-6)

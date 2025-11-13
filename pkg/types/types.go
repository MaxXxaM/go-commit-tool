package types

import "time"

// ModulePath представляет путь к Go модулю (например, github.com/user/repo)
type ModulePath string

// Version представляет версию модуля в формате semantic versioning
type Version string

// PlacementMode определяет где размещается локальная копия модуля
type PlacementMode string

const (
	// PlacementLocal - размещение рядом с текущим проектом
	PlacementLocal PlacementMode = "local"
	// PlacementShared - размещение в общей директории ~/shared
	PlacementShared PlacementMode = "shared"
	// PlacementAuto - автоматическое определение на основе количества зависимых модулей
	PlacementAuto PlacementMode = "auto"
)

// WorkspaceMode определяет режим работы с модулем
type WorkspaceMode string

const (
	// WorkspaceModeLocal - работа через go.work (локальная разработка)
	WorkspaceModeLocal WorkspaceMode = "workspace"
	// WorkspaceModeRemote - работа через go.mod (удаленные версии)
	WorkspaceModeRemote WorkspaceMode = "remote"
)

// VersionBump определяет тип изменения версии
type VersionBump string

const (
	VersionBumpMajor VersionBump = "major"
	VersionBumpMinor VersionBump = "minor"
	VersionBumpPatch VersionBump = "patch"
)

// Module представляет информацию о Go модуле
type Module struct {
	// Path - путь модуля (например, github.com/user/repo)
	Path ModulePath
	// Version - текущая версия модуля
	Version Version
	// LocalPath - локальный путь к модулю (если клонирован)
	LocalPath string
	// Placement - режим размещения (local/shared)
	Placement PlacementMode
	// Mode - режим работы (workspace/remote)
	Mode WorkspaceMode
	// HasChanges - есть ли незакоммиченные изменения
	HasChanges bool
	// Dependencies - прямые зависимости модуля
	Dependencies []*Module
	// Dependents - модули, которые зависят от этого
	Dependents []*Module
	// IsExternal - внешний модуль (не должен обрабатываться утилитой)
	IsExternal bool
	// LastUpdated - время последнего обновления метаданных
	LastUpdated time.Time
}

// DependencyTree представляет дерево зависимостей проекта
type DependencyTree struct {
	// Root - корневой модуль (текущий проект)
	Root *Module
	// Modules - карта всех модулей в дереве (путь -> модуль)
	Modules map[ModulePath]*Module
	// TopologicalOrder - порядок обработки модулей (от листьев к корню)
	TopologicalOrder []*Module
}

// Config представляет конфигурацию утилиты
type Config struct {
	// SharedDir - директория для shared зависимостей
	SharedDir string
	// DefaultPlacement - стратегия размещения по умолчанию
	DefaultPlacement PlacementMode
	// PlacementRules - правила для автоматического определения placement
	PlacementRules []PlacementRule
	// Auth - настройки аутентификации
	Auth map[string]AuthConfig
	// Versioning - настройки версионирования
	Versioning VersioningConfig
	// Release - настройки релиза
	Release ReleaseConfig
	// Exclude - паттерны исключения модулей
	Exclude []string
}

// PlacementRule определяет правило для автоматического определения placement
type PlacementRule struct {
	Pattern string
	Mode    PlacementMode
}

// AuthConfig содержит настройки аутентификации для Git хоста
type AuthConfig struct {
	Method   string // "ssh", "token", "credentials"
	TokenEnv string // имя переменной окружения с токеном
}

// VersioningConfig содержит настройки версионирования
type VersioningConfig struct {
	AutoDetect bool
}

// ReleaseConfig содержит настройки релиза
type ReleaseConfig struct {
	AutoPush      bool
	CreateChangelog bool
	RunTests      bool
}

// CustomReplace представляет кастомную replace директиву
type CustomReplace struct {
	Module      ModulePath
	Replacement string
}

// ModuleMetadata содержит метаданные о модуле для сохранения
type ModuleMetadata struct {
	Path            ModulePath    `json:"path"`
	Placement       PlacementMode `json:"placement"`
	LocalPath       string        `json:"local_path"`
	Mode            WorkspaceMode `json:"mode"`
	CurrentVersion  Version       `json:"current_version"`
	HasChanges      bool          `json:"has_changes"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// ProjectMetadata содержит все метаданные проекта
type ProjectMetadata struct {
	Modules        map[ModulePath]ModuleMetadata `json:"modules"`
	CustomReplaces []CustomReplace               `json:"custom_replaces"`
	LastSync       time.Time                     `json:"last_sync"`
}

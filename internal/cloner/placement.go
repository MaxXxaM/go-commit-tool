package cloner

import (
	"strings"

	"github.com/MaxXxaM/gomm/pkg/types"
)

// PlacementAnalyzer определяет стратегию размещения модуля
type PlacementAnalyzer struct {
	config *types.Config
}

// NewPlacementAnalyzer создает новый анализатор
func NewPlacementAnalyzer(config *types.Config) *PlacementAnalyzer {
	return &PlacementAnalyzer{
		config: config,
	}
}

// DeterminePlacement определяет режим размещения для модуля
func (a *PlacementAnalyzer) DeterminePlacement(module *types.Module, tree *types.DependencyTree) types.PlacementMode {
	// 1. Проверяем правила из конфигурации
	for _, rule := range a.config.PlacementRules {
		if matchPattern(rule.Pattern, string(module.Path)) {
			return rule.Mode
		}
	}

	// 2. Если режим не auto, используем default
	if a.config.DefaultPlacement != types.PlacementAuto {
		return a.config.DefaultPlacement
	}

	// 3. Автоматическое определение на основе количества зависимых модулей
	dependentCount := countDependents(module, tree)

	// Если на модуль ссылается больше одного модуля - shared
	if dependentCount > 1 {
		return types.PlacementShared
	}

	// Иначе local
	return types.PlacementLocal
}

// countDependents подсчитывает количество модулей, зависящих от данного
func countDependents(module *types.Module, tree *types.DependencyTree) int {
	if module == nil || tree == nil {
		return 0
	}

	// Получаем модуль из дерева для актуальной информации
	treeModule, ok := tree.Modules[module.Path]
	if !ok {
		return 0
	}

	return len(treeModule.Dependents)
}

// matchPattern проверяет соответствие module path паттерну
func matchPattern(pattern, modulePath string) bool {
	// Простая реализация wildcard matching
	if !strings.Contains(pattern, "*") {
		return pattern == modulePath
	}

	parts := strings.Split(pattern, "*")
	if len(parts) == 0 {
		return true
	}

	// Проверяем начало
	if !strings.HasPrefix(modulePath, parts[0]) {
		return false
	}
	modulePath = modulePath[len(parts[0]):]

	// Проверяем середину
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(modulePath, parts[i])
		if idx == -1 {
			return false
		}
		modulePath = modulePath[idx+len(parts[i]):]
	}

	// Проверяем конец
	if len(parts) > 1 {
		return strings.HasSuffix(modulePath, parts[len(parts)-1])
	}

	return true
}

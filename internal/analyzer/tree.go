package analyzer

import (
	"fmt"
	"path/filepath"

	"github.com/MaxXxaM/gomm/pkg/types"
)

// TreeBuilder строит дерево зависимостей
type TreeBuilder struct {
	parser   *ModFileParser
	rootPath string
}

// NewTreeBuilder создает новый builder
func NewTreeBuilder(parser *ModFileParser, rootPath string) *TreeBuilder {
	return &TreeBuilder{
		parser:   parser,
		rootPath: rootPath,
	}
}

// BuildTree строит дерево зависимостей начиная с корневого модуля
func (b *TreeBuilder) BuildTree() (*types.DependencyTree, error) {
	// Парсим корневой go.mod
	root, err := b.parser.ParseGoModFromDir(b.rootPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить корневой модуль: %w", err)
	}

	tree := &types.DependencyTree{
		Root:    root,
		Modules: make(map[types.ModulePath]*types.Module),
	}

	// Добавляем корень в карту модулей
	tree.Modules[root.Path] = root

	// Рекурсивно обрабатываем зависимости
	if err := b.processDependencies(root, tree, 0, 10); err != nil {
		return nil, err
	}

	// Строим топологический порядок
	if err := b.buildTopologicalOrder(tree); err != nil {
		return nil, fmt.Errorf("не удалось построить топологический порядок: %w", err)
	}

	return tree, nil
}

// processDependencies рекурсивно обрабатывает зависимости
func (b *TreeBuilder) processDependencies(module *types.Module, tree *types.DependencyTree, depth, maxDepth int) error {
	if depth >= maxDepth {
		return nil // ограничение глубины рекурсии
	}

	for i, dep := range module.Dependencies {
		// Проверяем, не обработан ли уже этот модуль
		if existing, ok := tree.Modules[dep.Path]; ok {
			// Модуль уже в дереве, заменяем ссылку
			module.Dependencies[i] = existing
			// Добавляем текущий модуль в зависимые
			existing.Dependents = append(existing.Dependents, module)
			continue
		}

		// Добавляем модуль в карту
		tree.Modules[dep.Path] = dep
		dep.Dependents = append(dep.Dependents, module)

		// Для внешних модулей не пытаемся загружать go.mod
		// В будущем здесь будет логика для загрузки локальных копий
		if dep.IsExternal {
			continue
		}

		// Попытка загрузить go.mod зависимости (если она локальная)
		if dep.LocalPath != "" {
			depGoModPath := filepath.Join(dep.LocalPath, "go.mod")
			depModule, err := b.parser.ParseGoMod(depGoModPath)
			if err != nil {
				// Не критичная ошибка, продолжаем
				continue
			}

			// Обновляем информацию о зависимости
			dep.Dependencies = depModule.Dependencies

			// Рекурсивно обрабатываем подзависимости
			if err := b.processDependencies(dep, tree, depth+1, maxDepth); err != nil {
				return err
			}
		}
	}

	return nil
}

// buildTopologicalOrder строит топологический порядок обработки модулей (от листьев к корню)
func (b *TreeBuilder) buildTopologicalOrder(tree *types.DependencyTree) error {
	visited := make(map[types.ModulePath]bool)
	tempMark := make(map[types.ModulePath]bool)
	order := make([]*types.Module, 0, len(tree.Modules))

	var visit func(*types.Module) error
	visit = func(module *types.Module) error {
		if tempMark[module.Path] {
			return fmt.Errorf("обнаружена циклическая зависимость в модуле %s", module.Path)
		}

		if visited[module.Path] {
			return nil
		}

		tempMark[module.Path] = true

		// Сначала посещаем зависимости
		for _, dep := range module.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		tempMark[module.Path] = false
		visited[module.Path] = true
		order = append(order, module)

		return nil
	}

	// Начинаем с корня
	if err := visit(tree.Root); err != nil {
		return err
	}

	// Посещаем все оставшиеся модули
	for _, module := range tree.Modules {
		if !visited[module.Path] {
			if err := visit(module); err != nil {
				return err
			}
		}
	}

	tree.TopologicalOrder = order
	return nil
}

// GetLeafModules возвращает модули-листья (на которые никто не ссылается кроме root)
func GetLeafModules(tree *types.DependencyTree) []*types.Module {
	leaves := make([]*types.Module, 0)

	for _, module := range tree.Modules {
		if module.Path == tree.Root.Path {
			continue
		}

		// Модуль является листом если на него никто не ссылается
		// или ссылается только корневой модуль
		if len(module.Dependents) == 0 || (len(module.Dependents) == 1 && module.Dependents[0].Path == tree.Root.Path) {
			leaves = append(leaves, module)
		}
	}

	return leaves
}

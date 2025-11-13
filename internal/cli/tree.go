package cli

import (
	"fmt"
	"strings"

	"github.com/MaxXxaM/gomm/internal/analyzer"
	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
)

var (
	treeLocalOnly bool
	treeDepth     int
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Визуализация дерева зависимостей",
	Long: `Отображает дерево зависимостей проекта в виде ASCII дерева.

Показывает:
- Иерархию зависимостей
- Версии модулей
- Локальные/удаленные модули`,
	RunE: runTree,
}

func init() {
	rootCmd.AddCommand(treeCmd)
	treeCmd.Flags().BoolVarP(&treeLocalOnly, "local-only", "l", false, "показывать только локальные модули")
	treeCmd.Flags().IntVarP(&treeDepth, "depth", "d", 10, "максимальная глубина дерева")
}

func runTree(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(config.ConfigFileName)
	if err != nil {
		log.Warning("Конфигурация не найдена, используются настройки по умолчанию")
		cfg = config.DefaultConfig()
	}

	// Создаем парсер и строим дерево
	parser := analyzer.NewModFileParser(cfg.Exclude)
	builder := analyzer.NewTreeBuilder(parser, ".")

	tree, err := builder.BuildTree()
	if err != nil {
		printError(fmt.Errorf("не удалось построить дерево: %w", err))
		return err
	}

	// Визуализация
	fmt.Printf("%s %s\n", tree.Root.Path, formatVersion(tree.Root.Version))
	printTree(tree.Root, "", tree, make(map[types.ModulePath]bool), 0)

	return nil
}

func printTree(module *types.Module, prefix string, tree *types.DependencyTree, visited map[types.ModulePath]bool, depth int) {
	if depth >= treeDepth {
		return
	}

	if len(module.Dependencies) == 0 {
		return
	}

	for i, dep := range module.Dependencies {
		isLast := i == len(module.Dependencies)-1

		// Пропускаем внешние модули если установлен флаг local-only
		if treeLocalOnly && dep.IsExternal {
			continue
		}

		// Символы для отрисовки дерева
		var connector, childPrefix string
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		} else {
			connector = "├── "
			childPrefix = prefix + "│   "
		}

		// Проверяем, не был ли модуль уже посещен (циклическая зависимость)
		symbol := ""
		if visited[dep.Path] {
			symbol = " (↻)"
		}

		// Формируем строку вывода
		modeSymbol := getModuleSymbol(dep)
		fmt.Printf("%s%s%s %s%s%s\n",
			prefix,
			connector,
			dep.Path,
			formatVersion(dep.Version),
			modeSymbol,
			symbol,
		)

		// Рекурсивно выводим зависимости (если модуль не был посещен)
		if !visited[dep.Path] {
			visited[dep.Path] = true
			printTree(dep, childPrefix, tree, visited, depth+1)
		}
	}
}

func formatVersion(version types.Version) string {
	if version == "" {
		return ""
	}
	return fmt.Sprintf("\033[36m%s\033[0m", version)
}

func getModuleSymbol(module *types.Module) string {
	if module.IsExternal {
		return ""
	}

	if module.Mode == types.WorkspaceModeLocal {
		return " \033[32m[local]\033[0m"
	}

	if module.LocalPath != "" {
		return " \033[33m[cloned]\033[0m"
	}

	return ""
}

// printTreeSimple печатает дерево в простом формате (без цветов)
func printTreeSimple(modules []*types.Module, indent int) {
	indentStr := strings.Repeat("  ", indent)
	for _, mod := range modules {
		fmt.Printf("%s- %s %s\n", indentStr, mod.Path, mod.Version)
		if len(mod.Dependencies) > 0 {
			printTreeSimple(mod.Dependencies, indent+1)
		}
	}
}

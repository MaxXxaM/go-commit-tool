package cli

import (
	"fmt"

	"github.com/MaxXxaM/gomm/internal/analyzer"
	"github.com/MaxXxaM/gomm/internal/cloner"
	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/internal/git"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
)

var (
	cloneMode string
	cloneAll  bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone [module-path...]",
	Short: "Клонирование зависимостей локально",
	Long: `Клонирует один или несколько модулей локально.

Режимы размещения:
  local  - рядом с текущим проектом
  shared - в общей директории (~/shared)
  auto   - автоматическое определение (по умолчанию)

Примеры:
  gomm clone github.com/user/repo
  gomm clone github.com/user/repo --mode local
  gomm clone --all`,
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVarP(&cloneMode, "mode", "m", "auto", "режим размещения (local, shared, auto)")
	cloneCmd.Flags().BoolVarP(&cloneAll, "all", "a", false, "клонировать все зависимости")
}

func runClone(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Проверяем git
	if !git.IsGitInstalled() {
		printError(fmt.Errorf("git не установлен"))
		return fmt.Errorf("git не установлен")
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(config.ConfigFileName)
	if err != nil {
		log.Warning("Конфигурация не найдена, используются настройки по умолчанию")
		cfg = config.DefaultConfig()
	}

	// Парсим режим размещения
	placement := types.PlacementMode(cloneMode)
	if placement != types.PlacementLocal && placement != types.PlacementShared && placement != types.PlacementAuto {
		printError(fmt.Errorf("неизвестный режим размещения: %s", cloneMode))
		return fmt.Errorf("неизвестный режим размещения: %s", cloneMode)
	}

	// Создаем cloner
	moduleCloner := cloner.NewModuleCloner(cfg, log)

	// Если --all, клонируем все зависимости
	if cloneAll {
		return cloneAllDependencies(moduleCloner, cfg, placement)
	}

	// Иначе клонируем указанные модули
	if len(args) == 0 {
		printError(fmt.Errorf("укажите module path или используйте --all"))
		return fmt.Errorf("укажите module path или используйте --all")
	}

	// Строим дерево для определения placement
	parser := analyzer.NewModFileParser(cfg.Exclude)
	builder := analyzer.NewTreeBuilder(parser, ".")
	tree, err := builder.BuildTree()
	if err != nil {
		log.Warning("Не удалось построить дерево зависимостей: %v", err)
		tree = nil
	}

	// Клонируем каждый указанный модуль
	for _, modulePath := range args {
		module := &types.Module{
			Path: types.ModulePath(modulePath),
		}

		// Ищем модуль в дереве если оно есть
		if tree != nil {
			if treeModule, ok := tree.Modules[module.Path]; ok {
				module = treeModule
			}
		}

		if err := moduleCloner.Clone(module, tree, placement); err != nil {
			printError(fmt.Errorf("не удалось клонировать %s: %w", modulePath, err))
			continue
		}
	}

	log.Success("Клонирование завершено")
	return nil
}

func cloneAllDependencies(moduleCloner *cloner.ModuleCloner, cfg *types.Config, placement types.PlacementMode) error {
	log := getLogger()

	log.Info("Сканирование зависимостей...")

	// Строим дерево зависимостей
	parser := analyzer.NewModFileParser(cfg.Exclude)
	builder := analyzer.NewTreeBuilder(parser, ".")
	tree, err := builder.BuildTree()
	if err != nil {
		printError(fmt.Errorf("не удалось построить дерево зависимостей: %w", err))
		return err
	}

	if len(tree.Root.Dependencies) == 0 {
		log.Info("Нет зависимостей для клонирования")
		return nil
	}

	log.Info("Найдено зависимостей: %d", len(tree.Root.Dependencies))

	// Клонируем каждую зависимость
	successCount := 0
	errorCount := 0

	for _, dep := range tree.Root.Dependencies {
		if dep.IsExternal {
			log.Debug("Пропускаем внешний модуль: %s", dep.Path)
			continue
		}

		log.Info("Клонирование %s...", dep.Path)

		if err := moduleCloner.Clone(dep, tree, placement); err != nil {
			log.ErrorMsg("Не удалось клонировать %s: %v", dep.Path, err)
			errorCount++
			continue
		}

		successCount++
	}

	fmt.Println()
	log.Success("Клонировано: %d модулей", successCount)
	if errorCount > 0 {
		log.Warning("Ошибок: %d", errorCount)
	}

	return nil
}

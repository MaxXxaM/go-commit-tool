package cli

import (
	"fmt"

	"github.com/MaxXxaM/gomm/internal/analyzer"
	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/spf13/cobra"
)

var (
	scanDepth int
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Сканирование зависимостей проекта",
	Long: `Сканирует go.mod и строит дерево зависимостей проекта.

Анализирует:
- Прямые зависимости
- Транзитивные зависимости (если доступны локально)
- Построение графа зависимостей`,
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().IntVarP(&scanDepth, "depth", "d", 10, "максимальная глубина сканирования")
}

func runScan(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(config.ConfigFileName)
	if err != nil {
		log.Warning("Конфигурация не найдена, используются настройки по умолчанию")
		cfg = config.DefaultConfig()
	}

	log.Info("Сканирование зависимостей...")

	// Создаем парсер
	parser := analyzer.NewModFileParser(cfg.Exclude)

	// Создаем builder дерева
	builder := analyzer.NewTreeBuilder(parser, ".")

	// Строим дерево
	tree, err := builder.BuildTree()
	if err != nil {
		printError(fmt.Errorf("не удалось построить дерево зависимостей: %w", err))
		return err
	}

	log.Success("Сканирование завершено")

	// Статистика
	fmt.Println()
	fmt.Printf("Корневой модуль: %s\n", tree.Root.Path)
	fmt.Printf("Всего модулей в дереве: %d\n", len(tree.Modules))
	fmt.Printf("Прямых зависимостей: %d\n", len(tree.Root.Dependencies))

	// Показываем прямые зависимости
	if len(tree.Root.Dependencies) > 0 {
		fmt.Println("\nПрямые зависимости:")
		for _, dep := range tree.Root.Dependencies {
			fmt.Printf("  - %s %s\n", dep.Path, dep.Version)
		}
	}

	fmt.Println("\nДля визуализации дерева выполните:")
	fmt.Println("  gomm tree")

	return nil
}

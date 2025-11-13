package cli

import (
	"fmt"

	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/internal/versioning"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Управление версиями модулей",
	Long:  `Команды для работы с версиями модулей.`,
}

var versionCheckCmd = &cobra.Command{
	Use:   "check [module-path]",
	Short: "Проверить текущую версию и предложить следующую",
	Long: `Анализирует изменения модуля и предлагает следующую версию.

Примеры:
  gomm version check
  gomm version check github.com/user/repo`,
	Args: cobra.MaximumNArgs(1),
	RunE: runVersionCheck,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.AddCommand(versionCheckCmd)
}

func runVersionCheck(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Определяем путь к модулю
	var modulePath string
	if len(args) > 0 {
		// Модуль указан, ищем его в метаданных
		modPath := types.ModulePath(args[0])

		metadataMgr := config.NewMetadataManager()
		metadata, err := metadataMgr.LoadOrCreate()
		if err != nil {
			printError(fmt.Errorf("не удалось загрузить метаданные: %w", err))
			return err
		}

		modMeta, ok := metadataMgr.GetModule(metadata, modPath)
		if !ok {
			printError(fmt.Errorf("модуль %s не найден в локальных копиях", modPath))
			return fmt.Errorf("модуль не найден")
		}

		modulePath = modMeta.LocalPath
	} else {
		// Используем текущую директорию
		modulePath = "."
	}

	// Создаем анализатор
	analyzer := versioning.NewAnalyzer(log)

	// Получаем текущую версию
	currentVersion, err := analyzer.GetCurrentVersion(modulePath)
	if err != nil {
		printError(fmt.Errorf("не удалось получить текущую версию: %w", err))
		return err
	}

	fmt.Println("=== Информация о версии ===")
	fmt.Println()
	fmt.Printf("Текущая версия: %s\n", currentVersion)
	fmt.Println()

	// Анализируем изменения
	bump, changes, err := analyzer.AnalyzeChanges(modulePath)
	if err != nil {
		printError(fmt.Errorf("не удалось проанализировать изменения: %w", err))
		return err
	}

	// Предлагаем следующую версию
	nextVersion, err := versioning.BumpVersion(currentVersion, bump)
	if err != nil {
		printError(fmt.Errorf("не удалось вычислить следующую версию: %w", err))
		return err
	}

	fmt.Printf("Предлагаемая версия: %s (%s)\n", nextVersion, bump)
	fmt.Println()

	if verbose && len(changes) > 0 {
		fmt.Println("Обнаруженные изменения:")
		for _, change := range changes {
			fmt.Printf("  - %s\n", change)
		}
		fmt.Println()
	}

	fmt.Println("Для создания релиза выполните:")
	if len(args) > 0 {
		fmt.Printf("  gomm release %s\n", args[0])
	} else {
		fmt.Println("  gomm release")
	}

	return nil
}

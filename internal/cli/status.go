package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Показать статус проекта и модулей",
	Long: `Отображает текущее состояние проекта:
- Конфигурация
- Локальные модули и их режимы
- Изменения в модулях`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Проверяем инициализацию
	if _, err := os.Stat(config.ConfigFileName); os.IsNotExist(err) {
		log.Warning("Проект не инициализирован")
		fmt.Println("Выполните: gomm init")
		return nil
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(config.ConfigFileName)
	if err != nil {
		printError(fmt.Errorf("не удалось загрузить конфигурацию: %w", err))
		return err
	}

	// Загружаем метаданные
	metadata, err := loadMetadata()
	if err != nil {
		log.Warning("Метаданные не найдены")
		fmt.Println("Выполните: gomm scan")
		return nil
	}

	// Вывод информации
	fmt.Println("=== Статус проекта ===")
	fmt.Println()

	fmt.Printf("Shared директория: %s\n", cfg.SharedDir)
	fmt.Printf("Стратегия размещения: %s\n", cfg.DefaultPlacement)
	fmt.Println()

	if len(metadata.Modules) == 0 {
		fmt.Println("Локальных модулей не найдено")
		fmt.Println("Для клонирования зависимостей выполните: gomm clone")
		return nil
	}

	fmt.Printf("Локальных модулей: %d\n", len(metadata.Modules))
	fmt.Println()

	// Группируем по режиму
	workspaceCount := 0
	remoteCount := 0
	changedCount := 0

	for _, mod := range metadata.Modules {
		if mod.Mode == types.WorkspaceModeLocal {
			workspaceCount++
		} else {
			remoteCount++
		}
		if mod.HasChanges {
			changedCount++
		}
	}

	fmt.Printf("В режиме workspace: %d\n", workspaceCount)
	fmt.Printf("В режиме remote: %d\n", remoteCount)
	fmt.Printf("С изменениями: %d\n", changedCount)

	if verbose {
		fmt.Println("\nПодробности:")
		for path, mod := range metadata.Modules {
			fmt.Printf("\n%s:\n", path)
			fmt.Printf("  Версия: %s\n", mod.CurrentVersion)
			fmt.Printf("  Режим: %s\n", mod.Mode)
			fmt.Printf("  Размещение: %s\n", mod.Placement)
			fmt.Printf("  Путь: %s\n", mod.LocalPath)
			if mod.HasChanges {
				fmt.Println("  ⚠ Есть незакоммиченные изменения")
			}
		}
	}

	return nil
}

func loadMetadata() (*types.ProjectMetadata, error) {
	metadataPath := config.GetMetadataPath()

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata types.ProjectMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("не удалось распарсить метаданные: %w", err)
	}

	return &metadata, nil
}

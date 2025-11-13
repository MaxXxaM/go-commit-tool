package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Инициализация gomm в текущей директории",
	Long: `Инициализирует gomm в текущей директории проекта.

Создает:
- Конфигурационный файл .gomm.yaml
- Директорию .gomm/ для метаданных
- Сканирует зависимости из go.mod`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Проверяем наличие go.mod
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		printError(fmt.Errorf("go.mod не найден в текущей директории"))
		return err
	}

	log.Info("Инициализация gomm...")

	// Проверяем, не инициализирован ли уже проект
	if _, err := os.Stat(config.ConfigFileName); err == nil {
		log.Warning("Проект уже инициализирован (найден %s)", config.ConfigFileName)
		return nil
	}

	// Создаем директорию для метаданных
	metadataDir := filepath.Join(".", config.MetadataDir)
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		printError(fmt.Errorf("не удалось создать директорию %s: %w", metadataDir, err))
		return err
	}

	// Создаем конфигурацию по умолчанию
	cfg := config.DefaultConfig()
	if err := config.SaveConfig(cfg, config.ConfigFileName); err != nil {
		printError(fmt.Errorf("не удалось создать конфигурационный файл: %w", err))
		return err
	}

	log.Success("Проект инициализирован")
	log.Info("Создан файл конфигурации: %s", config.ConfigFileName)
	log.Info("Создана директория метаданных: %s", metadataDir)

	fmt.Println("\nДля сканирования зависимостей выполните:")
	fmt.Println("  gomm scan")

	return nil
}

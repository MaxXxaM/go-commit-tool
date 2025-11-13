package cli

import (
	"fmt"

	"github.com/MaxXxaM/gomm/internal/cloner"
	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
)

var (
	cleanForce bool
	cleanAll   bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean [module-path...]",
	Short: "Удаление локальных копий модулей",
	Long: `Удаляет локальные копии указанных модулей.

По умолчанию проверяет наличие незакоммиченных изменений.
Используйте --force для принудительного удаления.

Примеры:
  gomm clean github.com/user/repo
  gomm clean github.com/user/repo --force
  gomm clean --all`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "принудительное удаление без проверки изменений")
	cleanCmd.Flags().BoolVarP(&cleanAll, "all", "a", false, "удалить все локальные копии")
}

func runClean(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Проверяем аргументы
	if !cleanAll && len(args) == 0 {
		printError(fmt.Errorf("укажите module path или используйте --all"))
		return fmt.Errorf("укажите module path или используйте --all")
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(config.ConfigFileName)
	if err != nil {
		printError(fmt.Errorf("не удалось загрузить конфигурацию: %w", err))
		return err
	}

	// Загружаем метаданные
	metadataMgr := config.NewMetadataManager()
	metadata, err := metadataMgr.LoadOrCreate()
	if err != nil {
		printError(fmt.Errorf("не удалось загрузить метаданные: %w", err))
		return err
	}

	// Создаем cloner
	moduleCloner := cloner.NewModuleCloner(cfg, log)

	if cleanAll {
		return cleanAllModules(moduleCloner, metadata, metadataMgr)
	}

	// Удаляем указанные модули
	for _, modulePath := range args {
		if err := cleanModule(moduleCloner, types.ModulePath(modulePath), metadata, metadataMgr); err != nil {
			log.ErrorMsg("Не удалось удалить %s: %v", modulePath, err)
			continue
		}
	}

	return nil
}

func cleanModule(moduleCloner *cloner.ModuleCloner, modulePath types.ModulePath, metadata *types.ProjectMetadata, metadataMgr *config.MetadataManager) error {
	log := getLogger()

	// Ищем модуль в метаданных
	modMeta, ok := metadataMgr.GetModule(metadata, modulePath)
	if !ok {
		return fmt.Errorf("модуль %s не найден в локальных копиях", modulePath)
	}

	// Создаем модуль из метаданных
	module := &types.Module{
		Path:       modMeta.Path,
		Version:    modMeta.CurrentVersion,
		LocalPath:  modMeta.LocalPath,
		Placement:  modMeta.Placement,
		Mode:       modMeta.Mode,
		HasChanges: modMeta.HasChanges,
	}

	// Удаляем модуль
	if err := moduleCloner.Clean(module, cleanForce); err != nil {
		return err
	}

	log.Success("Модуль %s удален", modulePath)
	return nil
}

func cleanAllModules(moduleCloner *cloner.ModuleCloner, metadata *types.ProjectMetadata, metadataMgr *config.MetadataManager) error {
	log := getLogger()

	if len(metadata.Modules) == 0 {
		log.Info("Нет локальных модулей для удаления")
		return nil
	}

	log.Info("Найдено локальных модулей: %d", len(metadata.Modules))

	if !cleanForce {
		log.Warning("Это удалит все локальные копии модулей!")
		fmt.Print("Продолжить? (y/N): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			log.Info("Отменено")
			return nil
		}
	}

	successCount := 0
	errorCount := 0

	// Копируем ключи для итерации (так как будем удалять из map)
	modulePaths := make([]types.ModulePath, 0, len(metadata.Modules))
	for path := range metadata.Modules {
		modulePaths = append(modulePaths, path)
	}

	for _, path := range modulePaths {
		if err := cleanModule(moduleCloner, path, metadata, metadataMgr); err != nil {
			log.ErrorMsg("Не удалось удалить %s: %v", path, err)
			errorCount++
			continue
		}
		successCount++
	}

	fmt.Println()
	log.Success("Удалено: %d модулей", successCount)
	if errorCount > 0 {
		log.Warning("Ошибок: %d", errorCount)
	}

	return nil
}

package cli

import (
	"fmt"
	"strings"

	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/internal/workspace"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
)

var (
	localAll    bool
	localPrefix string
)

var localCmd = &cobra.Command{
	Use:   "local [module-path...]",
	Short: "Переключение модулей в режим локальной разработки",
	Long: `Переключает один или несколько модулей в режим workspace (go.work).

Модули должны быть предварительно клонированы локально.

Примеры:
  gomm local github.com/user/repo
  gomm local github.com/user/repo1 github.com/user/repo2
  gomm local --all
  gomm local --prefix github.com/myorg`,
	RunE: runLocal,
}

func init() {
	rootCmd.AddCommand(localCmd)
	localCmd.Flags().BoolVarP(&localAll, "all", "a", false, "переключить все клонированные модули")
	localCmd.Flags().StringVarP(&localPrefix, "prefix", "p", "", "переключить модули по префиксу")
}

func runLocal(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Проверяем аргументы
	if !localAll && localPrefix == "" && len(args) == 0 {
		printError(fmt.Errorf("укажите module path, используйте --all или --prefix"))
		return fmt.Errorf("укажите module path, используйте --all или --prefix")
	}

	// Загружаем метаданные
	metadataMgr := config.NewMetadataManager()
	metadata, err := metadataMgr.LoadOrCreate()
	if err != nil {
		printError(fmt.Errorf("не удалось загрузить метаданные: %w", err))
		return err
	}

	if len(metadata.Modules) == 0 {
		log.Warning("Нет клонированных модулей")
		fmt.Println("Для клонирования выполните: gomm clone")
		return nil
	}

	// Создаем workspace manager
	wsMgr := workspace.NewManager(log)

	// Определяем модули для переключения
	modulesToSwitch := make([]*types.Module, 0)

	if localAll {
		// Все модули
		for path, modMeta := range metadata.Modules {
			module := &types.Module{
				Path:       path,
				Version:    modMeta.CurrentVersion,
				LocalPath:  modMeta.LocalPath,
				Placement:  modMeta.Placement,
				Mode:       modMeta.Mode,
				HasChanges: modMeta.HasChanges,
			}
			modulesToSwitch = append(modulesToSwitch, module)
		}
	} else if localPrefix != "" {
		// По префиксу
		for path, modMeta := range metadata.Modules {
			if strings.HasPrefix(string(path), localPrefix) {
				module := &types.Module{
					Path:       path,
					Version:    modMeta.CurrentVersion,
					LocalPath:  modMeta.LocalPath,
					Placement:  modMeta.Placement,
					Mode:       modMeta.Mode,
					HasChanges: modMeta.HasChanges,
				}
				modulesToSwitch = append(modulesToSwitch, module)
			}
		}
	} else {
		// Конкретные модули
		for _, modulePath := range args {
			modPath := types.ModulePath(modulePath)
			modMeta, ok := metadataMgr.GetModule(metadata, modPath)
			if !ok {
				log.Warning("Модуль %s не найден в локальных копиях", modulePath)
				continue
			}

			module := &types.Module{
				Path:       modPath,
				Version:    modMeta.CurrentVersion,
				LocalPath:  modMeta.LocalPath,
				Placement:  modMeta.Placement,
				Mode:       modMeta.Mode,
				HasChanges: modMeta.HasChanges,
			}
			modulesToSwitch = append(modulesToSwitch, module)
		}
	}

	if len(modulesToSwitch) == 0 {
		log.Warning("Нет модулей для переключения")
		return nil
	}

	log.Info("Переключение %d модулей в режим workspace...", len(modulesToSwitch))

	// Переключаем модули
	successCount := 0
	for _, module := range modulesToSwitch {
		// Пропускаем модули уже в workspace режиме
		if module.Mode == types.WorkspaceModeLocal {
			log.Debug("Модуль %s уже в workspace режиме", module.Path)
			continue
		}

		// Добавляем в go.work
		if err := wsMgr.AddModule(module); err != nil {
			log.ErrorMsg("Не удалось добавить %s в go.work: %v", module.Path, err)
			continue
		}

		// Обновляем метаданные
		if err := metadataMgr.UpdateModule(metadata, module.Path, func(m *types.ModuleMetadata) {
			m.Mode = types.WorkspaceModeLocal
		}); err != nil {
			log.ErrorMsg("Не удалось обновить метаданные для %s: %v", module.Path, err)
			continue
		}

		log.Success("Модуль %s переключен в workspace режим", module.Path)
		successCount++
	}

	// Сохраняем метаданные
	if err := metadataMgr.Save(metadata); err != nil {
		log.Warning("Не удалось сохранить метаданные: %v", err)
	}

	fmt.Println()
	log.Success("Переключено модулей: %d", successCount)

	if wsMgr.Exists() {
		log.Info("go.work файл обновлен")
		log.Info("Теперь Go будет использовать локальные версии модулей")
	}

	return nil
}

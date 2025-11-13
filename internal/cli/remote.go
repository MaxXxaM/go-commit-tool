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
	remoteAll    bool
	remotePrefix string
)

var remoteCmd = &cobra.Command{
	Use:   "remote [module-path...]",
	Short: "Переключение модулей на удаленные версии",
	Long: `Переключает один или несколько модулей обратно в режим go.mod (удаленные версии).

Удаляет модули из go.work и Go начнет использовать версии из go.mod.

Примеры:
  gomm remote github.com/user/repo
  gomm remote github.com/user/repo1 github.com/user/repo2
  gomm remote --all
  gomm remote --prefix github.com/myorg`,
	RunE: runRemote,
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.Flags().BoolVarP(&remoteAll, "all", "a", false, "переключить все модули на remote")
	remoteCmd.Flags().StringVarP(&remotePrefix, "prefix", "p", "", "переключить модули по префиксу")
}

func runRemote(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Проверяем аргументы
	if !remoteAll && remotePrefix == "" && len(args) == 0 {
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
		log.Info("Нет клонированных модулей")
		return nil
	}

	// Создаем workspace manager
	wsMgr := workspace.NewManager(log)

	if !wsMgr.Exists() && remoteAll {
		log.Info("go.work не существует, все модули уже в remote режиме")
		return nil
	}

	// Определяем модули для переключения
	modulesToSwitch := make([]*types.Module, 0)

	if remoteAll {
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
	} else if remotePrefix != "" {
		// По префиксу
		for path, modMeta := range metadata.Modules {
			if strings.HasPrefix(string(path), remotePrefix) {
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

	log.Info("Переключение %d модулей в режим remote...", len(modulesToSwitch))

	// Переключаем модули
	successCount := 0
	for _, module := range modulesToSwitch {
		// Пропускаем модули уже в remote режиме
		if module.Mode == types.WorkspaceModeRemote {
			log.Debug("Модуль %s уже в remote режиме", module.Path)
			continue
		}

		// Удаляем из go.work
		if wsMgr.Exists() {
			if err := wsMgr.RemoveModule(module); err != nil {
				log.ErrorMsg("Не удалось удалить %s из go.work: %v", module.Path, err)
				continue
			}
		}

		// Обновляем метаданные
		if err := metadataMgr.UpdateModule(metadata, module.Path, func(m *types.ModuleMetadata) {
			m.Mode = types.WorkspaceModeRemote
		}); err != nil {
			log.ErrorMsg("Не удалось обновить метаданные для %s: %v", module.Path, err)
			continue
		}

		log.Success("Модуль %s переключен в remote режим", module.Path)
		successCount++
	}

	// Сохраняем метаданные
	if err := metadataMgr.Save(metadata); err != nil {
		log.Warning("Не удалось сохранить метаданные: %v", err)
	}

	fmt.Println()
	log.Success("Переключено модулей: %d", successCount)

	if !wsMgr.Exists() {
		log.Info("go.work удален, Go будет использовать версии из go.mod")
	} else {
		log.Info("go.work обновлен")
	}

	return nil
}

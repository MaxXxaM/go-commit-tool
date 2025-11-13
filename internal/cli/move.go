package cli

import (
	"fmt"

	"github.com/MaxXxaM/gomm/internal/cloner"
	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
)

var (
	moveTo string
)

var moveCmd = &cobra.Command{
	Use:   "move <module-path> --to local|shared",
	Short: "Перемещение модуля между local и shared режимами",
	Long: `Перемещает локальную копию модуля между режимами размещения.

Режимы:
  local  - рядом с текущим проектом
  shared - в общей директории (~/shared)

Примеры:
  gomm move github.com/user/repo --to shared
  gomm move github.com/user/lib --to local`,
	Args: cobra.ExactArgs(1),
	RunE: runMove,
}

func init() {
	rootCmd.AddCommand(moveCmd)
	moveCmd.Flags().StringVar(&moveTo, "to", "", "целевой режим размещения (local, shared)")
	moveCmd.MarkFlagRequired("to")
}

func runMove(cmd *cobra.Command, args []string) error {
	log := getLogger()

	modulePath := types.ModulePath(args[0])

	// Проверяем целевой режим
	targetPlacement := types.PlacementMode(moveTo)
	if targetPlacement != types.PlacementLocal && targetPlacement != types.PlacementShared {
		printError(fmt.Errorf("неизвестный режим размещения: %s (используйте local или shared)", moveTo))
		return fmt.Errorf("неизвестный режим размещения: %s", moveTo)
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

	// Ищем модуль в метаданных
	modMeta, ok := metadataMgr.GetModule(metadata, modulePath)
	if !ok {
		printError(fmt.Errorf("модуль %s не найден в локальных копиях", modulePath))
		fmt.Println("Для клонирования выполните: gomm clone", modulePath)
		return fmt.Errorf("модуль не найден")
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

	// Создаем cloner
	moduleCloner := cloner.NewModuleCloner(cfg, log)

	// Перемещаем модуль
	log.Info("Перемещение %s в режим %s...", modulePath, targetPlacement)

	if err := moduleCloner.Move(module, targetPlacement); err != nil {
		printError(err)
		return err
	}

	log.Success("Модуль перемещен в режим %s", targetPlacement)
	log.Info("Новый путь: %s", module.LocalPath)

	return nil
}

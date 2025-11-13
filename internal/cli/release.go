package cli

import (
	"fmt"

	"github.com/MaxXxaM/gomm/internal/analyzer"
	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/internal/versioning"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
)

var (
	releaseBump    string
	releaseVersion string
	releaseAuto    bool
	releaseNoTests bool
	releaseCascade bool
)

var releaseCmd = &cobra.Command{
	Use:   "release [module-path]",
	Short: "Создание релиза модуля",
	Long: `Создает релиз (тег) для модуля и опционально обновляет зависимые модули.

Примеры:
  gomm release                                    # текущий модуль, auto режим
  gomm release --auto                             # автоопределение версии
  gomm release --bump patch                       # patch версия
  gomm release --version v1.2.3                   # конкретная версия
  gomm release github.com/user/repo --cascade     # каскадный релиз`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRelease,
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.Flags().StringVar(&releaseBump, "bump", "", "тип версии (major, minor, patch)")
	releaseCmd.Flags().StringVar(&releaseVersion, "version", "", "конкретная версия (v1.2.3)")
	releaseCmd.Flags().BoolVar(&releaseAuto, "auto", true, "автоматическое определение версии")
	releaseCmd.Flags().BoolVar(&releaseNoTests, "no-tests", false, "пропустить запуск тестов")
	releaseCmd.Flags().BoolVar(&releaseCascade, "cascade", false, "каскадный релиз зависимых модулей")
}

func runRelease(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(config.ConfigFileName)
	if err != nil {
		log.Warning("Конфигурация не найдена, используются настройки по умолчанию")
		cfg = config.DefaultConfig()
	}

	// Определяем модуль
	var module *types.Module
	var modulePath string

	if len(args) > 0 {
		// Модуль указан
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

		module = &types.Module{
			Path:       modPath,
			LocalPath:  modMeta.LocalPath,
			Version:    modMeta.CurrentVersion,
			Placement:  modMeta.Placement,
			Mode:       modMeta.Mode,
			HasChanges: modMeta.HasChanges,
		}
		modulePath = modMeta.LocalPath
	} else {
		// Используем текущую директорию
		modulePath = "."

		// Парсим go.mod для получения module path
		parser := analyzer.NewModFileParser(cfg.Exclude)
		mod, err := parser.ParseGoModFromDir(modulePath)
		if err != nil {
			printError(fmt.Errorf("не удалось распарсить go.mod: %w", err))
			return err
		}

		module = mod
		module.LocalPath = modulePath
	}

	// Создаем release manager
	releaseMgr := versioning.NewReleaseManager(cfg, log)
	versionAnalyzer := versioning.NewAnalyzer(log)

	// Определяем версию
	var newVersion types.Version

	if releaseVersion != "" {
		// Конкретная версия указана
		newVersion = types.Version(releaseVersion)
		if err := versioning.ValidateVersion(newVersion); err != nil {
			printError(fmt.Errorf("невалидная версия: %w", err))
			return err
		}
	} else if releaseBump != "" {
		// Тип bump указан
		bump := types.VersionBump(releaseBump)
		if bump != types.VersionBumpMajor && bump != types.VersionBumpMinor && bump != types.VersionBumpPatch {
			printError(fmt.Errorf("невалидный тип bump: %s (используйте major, minor или patch)", releaseBump))
			return fmt.Errorf("невалидный тип bump")
		}

		currentVersion, err := versionAnalyzer.GetCurrentVersion(modulePath)
		if err != nil {
			printError(fmt.Errorf("не удалось получить текущую версию: %w", err))
			return err
		}

		newVersion, err = versioning.BumpVersion(currentVersion, bump)
		if err != nil {
			printError(fmt.Errorf("не удалось вычислить новую версию: %w", err))
			return err
		}
	} else {
		// Автоматическое определение
		suggestedVersion, bump, _, err := versionAnalyzer.SuggestNextVersion(modulePath)
		if err != nil {
			printError(fmt.Errorf("не удалось определить версию: %w", err))
			return err
		}

		newVersion = suggestedVersion
		log.Info("Автоопределенная версия: %s (%s)", newVersion, bump)
	}

	// Создаем релиз
	runTests := cfg.Release.RunTests && !releaseNoTests

	log.Info("Создание релиза %s для %s...", newVersion, module.Path)

	if err := releaseMgr.Release(module, newVersion, runTests); err != nil {
		printError(err)
		return err
	}

	log.Success("Релиз %s успешно создан!", newVersion)
	log.Info("Модуль: %s", module.Path)
	log.Info("Версия: %s", newVersion)

	if cfg.Release.AutoPush {
		log.Info("Тег запушен в remote репозиторий")
	} else {
		log.Warning("Для публикации тега выполните: git push origin %s", newVersion)
	}

	// Каскадный релиз если требуется
	if releaseCascade {
		log.Info("Запуск каскадного релиза...")

		// Строим дерево зависимостей
		parser := analyzer.NewModFileParser(cfg.Exclude)
		builder := analyzer.NewTreeBuilder(parser, ".")
		tree, err := builder.BuildTree()
		if err != nil {
			log.Warning("Не удалось построить дерево зависимостей для каскадного релиза: %v", err)
			return nil
		}

		// Запускаем каскадный релиз для зависимых модулей
		dependents := make([]*types.Module, 0)
		for _, dep := range tree.Modules {
			// Проверяем, зависит ли модуль от релизнутого
			for _, dependency := range dep.Dependencies {
				if dependency.Path == module.Path {
					dependents = append(dependents, dep)
					break
				}
			}
		}

		if len(dependents) > 0 {
			log.Info("Найдено зависимых модулей: %d", len(dependents))
			// TODO: Реализовать каскадное обновление
			log.Warning("Каскадное обновление зависимых модулей пока не реализовано")
		} else {
			log.Info("Нет зависимых модулей для обновления")
		}
	}

	return nil
}

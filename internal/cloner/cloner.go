package cloner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/internal/git"
	"github.com/MaxXxaM/gomm/pkg/logger"
	"github.com/MaxXxaM/gomm/pkg/types"
)

// ModuleCloner клонирует и управляет локальными копиями модулей
type ModuleCloner struct {
	gitClient         *git.Client
	placementAnalyzer *PlacementAnalyzer
	metadataManager   *config.MetadataManager
	config            *types.Config
	log               *logger.Logger
}

// NewModuleCloner создает новый cloner
func NewModuleCloner(cfg *types.Config, log *logger.Logger) *ModuleCloner {
	return &ModuleCloner{
		gitClient:         git.NewClient(log),
		placementAnalyzer: NewPlacementAnalyzer(cfg),
		metadataManager:   config.NewMetadataManager(),
		config:            cfg,
		log:               log,
	}
}

// Clone клонирует модуль локально
func (c *ModuleCloner) Clone(module *types.Module, tree *types.DependencyTree, forcePlacement types.PlacementMode) error {
	// Определяем стратегию размещения
	placement := forcePlacement
	if placement == "" || placement == types.PlacementAuto {
		placement = c.placementAnalyzer.DeterminePlacement(module, tree)
	}

	// Определяем целевую директорию
	targetDir, err := c.getTargetDirectory(module.Path, placement)
	if err != nil {
		return fmt.Errorf("не удалось определить целевую директорию: %w", err)
	}

	// Проверяем, не существует ли уже директория
	if _, err := os.Stat(targetDir); err == nil {
		// Проверяем, является ли это git репозиторием
		if c.gitClient.IsRepo(targetDir) {
			c.log.Info("Модуль %s уже клонирован в %s", module.Path, targetDir)

			// Обновляем метаданные
			return c.updateMetadata(module, targetDir, placement)
		}
		return fmt.Errorf("директория %s уже существует, но не является git репозиторием", targetDir)
	}

	// Конвертируем module path в repository URL
	repoURL := git.ModulePathToRepoURL(string(module.Path))

	c.log.Info("Клонирование %s в %s...", module.Path, targetDir)

	// Клонируем репозиторий
	if err := c.gitClient.Clone(repoURL, targetDir); err != nil {
		return fmt.Errorf("не удалось клонировать репозиторий: %w", err)
	}

	c.log.Success("Модуль клонирован: %s", targetDir)

	// Получаем информацию о версии
	currentTag, _ := c.gitClient.GetCurrentTag(targetDir)
	if currentTag != "" {
		module.Version = types.Version(currentTag)
	}

	// Обновляем информацию о модуле
	module.LocalPath = targetDir
	module.Placement = placement
	module.Mode = types.WorkspaceModeRemote // По умолчанию remote

	// Сохраняем метаданные
	return c.updateMetadata(module, targetDir, placement)
}

// Move перемещает модуль между local и shared режимами
func (c *ModuleCloner) Move(module *types.Module, newPlacement types.PlacementMode) error {
	if module.LocalPath == "" {
		return fmt.Errorf("модуль %s не клонирован локально", module.Path)
	}

	if module.Placement == newPlacement {
		c.log.Info("Модуль уже находится в режиме %s", newPlacement)
		return nil
	}

	// Проверяем наличие изменений
	hasChanges, err := c.gitClient.HasUncommittedChanges(module.LocalPath)
	if err != nil {
		return fmt.Errorf("не удалось проверить изменения: %w", err)
	}

	if hasChanges {
		return fmt.Errorf("модуль имеет незакоммиченные изменения, закоммитьте или отмените их")
	}

	// Определяем новую директорию
	newDir, err := c.getTargetDirectory(module.Path, newPlacement)
	if err != nil {
		return fmt.Errorf("не удалось определить целевую директорию: %w", err)
	}

	c.log.Info("Перемещение %s из %s в %s...", module.Path, module.LocalPath, newDir)

	// Создаем родительскую директорию
	parentDir := filepath.Dir(newDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", parentDir, err)
	}

	// Перемещаем директорию
	if err := os.Rename(module.LocalPath, newDir); err != nil {
		return fmt.Errorf("не удалось переместить директорию: %w", err)
	}

	c.log.Success("Модуль перемещен в %s", newDir)

	// Обновляем модуль
	oldPath := module.LocalPath
	module.LocalPath = newDir
	module.Placement = newPlacement

	// Обновляем метаданные
	if err := c.updateMetadata(module, newDir, newPlacement); err != nil {
		// Пытаемся откатить перемещение
		os.Rename(newDir, oldPath)
		return fmt.Errorf("не удалось обновить метаданные: %w", err)
	}

	return nil
}

// Clean удаляет локальную копию модуля
func (c *ModuleCloner) Clean(module *types.Module, force bool) error {
	if module.LocalPath == "" {
		return fmt.Errorf("модуль %s не клонирован локально", module.Path)
	}

	// Проверяем наличие изменений если не force
	if !force {
		hasChanges, err := c.gitClient.HasUncommittedChanges(module.LocalPath)
		if err != nil {
			return fmt.Errorf("не удалось проверить изменения: %w", err)
		}

		if hasChanges {
			return fmt.Errorf("модуль имеет незакоммиченные изменения, используйте --force для удаления")
		}
	}

	c.log.Info("Удаление локальной копии %s...", module.LocalPath)

	// Удаляем директорию
	if err := os.RemoveAll(module.LocalPath); err != nil {
		return fmt.Errorf("не удалось удалить директорию: %w", err)
	}

	c.log.Success("Локальная копия удалена")

	// Удаляем из метаданных
	metadata, err := c.metadataManager.LoadOrCreate()
	if err != nil {
		return fmt.Errorf("не удалось загрузить метаданные: %w", err)
	}

	c.metadataManager.RemoveModule(metadata, module.Path)

	if err := c.metadataManager.Save(metadata); err != nil {
		return fmt.Errorf("не удалось сохранить метаданные: %w", err)
	}

	return nil
}

// getTargetDirectory определяет целевую директорию для клонирования
func (c *ModuleCloner) getTargetDirectory(modulePath types.ModulePath, placement types.PlacementMode) (string, error) {
	// Извлекаем имя репозитория из module path
	// Например: github.com/user/repo -> repo
	parts := strings.Split(string(modulePath), "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("невалидный module path: %s", modulePath)
	}

	repoName := parts[len(parts)-1]

	switch placement {
	case types.PlacementLocal:
		// Размещаем рядом с текущим проектом
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("не удалось получить текущую директорию: %w", err)
		}
		parentDir := filepath.Dir(cwd)
		return filepath.Join(parentDir, repoName), nil

	case types.PlacementShared:
		// Размещаем в shared директории
		return filepath.Join(c.config.SharedDir, repoName), nil

	default:
		return "", fmt.Errorf("неизвестный режим размещения: %s", placement)
	}
}

// updateMetadata обновляет метаданные модуля
func (c *ModuleCloner) updateMetadata(module *types.Module, localPath string, placement types.PlacementMode) error {
	metadata, err := c.metadataManager.LoadOrCreate()
	if err != nil {
		return err
	}

	// Проверяем наличие изменений
	hasChanges := false
	if c.gitClient.IsRepo(localPath) {
		hasChanges, _ = c.gitClient.HasUncommittedChanges(localPath)
	}

	module.HasChanges = hasChanges
	module.LocalPath = localPath
	module.Placement = placement

	c.metadataManager.AddModule(metadata, module)

	return c.metadataManager.Save(metadata)
}

// UpdateFromGit обновляет информацию о модуле из git
func (c *ModuleCloner) UpdateFromGit(module *types.Module) error {
	if module.LocalPath == "" {
		return fmt.Errorf("модуль не имеет локального пути")
	}

	if !c.gitClient.IsRepo(module.LocalPath) {
		return fmt.Errorf("директория не является git репозиторием")
	}

	// Получаем текущий тег
	tag, err := c.gitClient.GetCurrentTag(module.LocalPath)
	if err == nil && tag != "" {
		module.Version = types.Version(tag)
	}

	// Проверяем изменения
	hasChanges, err := c.gitClient.HasUncommittedChanges(module.LocalPath)
	if err == nil {
		module.HasChanges = hasChanges
	}

	return nil
}

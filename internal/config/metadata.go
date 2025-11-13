package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/MaxXxaM/gomm/pkg/types"
)

// MetadataManager управляет метаданными проекта
type MetadataManager struct {
	filePath string
}

// NewMetadataManager создает новый менеджер метаданных
func NewMetadataManager() *MetadataManager {
	return &MetadataManager{
		filePath: GetMetadataPath(),
	}
}

// Load загружает метаданные из файла
func (m *MetadataManager) Load() (*types.ProjectMetadata, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Файл не существует, возвращаем пустые метаданные
			return m.createEmpty(), nil
		}
		return nil, fmt.Errorf("не удалось прочитать метаданные: %w", err)
	}

	var metadata types.ProjectMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("не удалось распарсить метаданные: %w", err)
	}

	return &metadata, nil
}

// Save сохраняет метаданные в файл
func (m *MetadataManager) Save(metadata *types.ProjectMetadata) error {
	metadata.LastSync = time.Now()

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось сериализовать метаданные: %w", err)
	}

	if err := os.WriteFile(m.filePath, data, 0644); err != nil {
		return fmt.Errorf("не удалось сохранить метаданные: %w", err)
	}

	return nil
}

// AddModule добавляет или обновляет модуль в метаданных
func (m *MetadataManager) AddModule(metadata *types.ProjectMetadata, module *types.Module) {
	if metadata.Modules == nil {
		metadata.Modules = make(map[types.ModulePath]types.ModuleMetadata)
	}

	metadata.Modules[module.Path] = types.ModuleMetadata{
		Path:           module.Path,
		Placement:      module.Placement,
		LocalPath:      module.LocalPath,
		Mode:           module.Mode,
		CurrentVersion: module.Version,
		HasChanges:     module.HasChanges,
		LastUpdated:    time.Now(),
	}
}

// RemoveModule удаляет модуль из метаданных
func (m *MetadataManager) RemoveModule(metadata *types.ProjectMetadata, modulePath types.ModulePath) {
	if metadata.Modules != nil {
		delete(metadata.Modules, modulePath)
	}
}

// GetModule возвращает метаданные модуля
func (m *MetadataManager) GetModule(metadata *types.ProjectMetadata, modulePath types.ModulePath) (types.ModuleMetadata, bool) {
	if metadata.Modules == nil {
		return types.ModuleMetadata{}, false
	}

	mod, ok := metadata.Modules[modulePath]
	return mod, ok
}

// UpdateModule обновляет метаданные модуля
func (m *MetadataManager) UpdateModule(metadata *types.ProjectMetadata, modulePath types.ModulePath, updateFn func(*types.ModuleMetadata)) error {
	if metadata.Modules == nil {
		return fmt.Errorf("модуль %s не найден в метаданных", modulePath)
	}

	mod, ok := metadata.Modules[modulePath]
	if !ok {
		return fmt.Errorf("модуль %s не найден в метаданных", modulePath)
	}

	updateFn(&mod)
	mod.LastUpdated = time.Now()
	metadata.Modules[modulePath] = mod

	return nil
}

// AddCustomReplace добавляет кастомную replace директиву
func (m *MetadataManager) AddCustomReplace(metadata *types.ProjectMetadata, replace types.CustomReplace) {
	if metadata.CustomReplaces == nil {
		metadata.CustomReplaces = make([]types.CustomReplace, 0)
	}

	// Проверяем, не существует ли уже
	for i, r := range metadata.CustomReplaces {
		if r.Module == replace.Module {
			metadata.CustomReplaces[i] = replace
			return
		}
	}

	metadata.CustomReplaces = append(metadata.CustomReplaces, replace)
}

// RemoveCustomReplace удаляет кастомную replace директиву
func (m *MetadataManager) RemoveCustomReplace(metadata *types.ProjectMetadata, modulePath types.ModulePath) {
	if metadata.CustomReplaces == nil {
		return
	}

	for i, r := range metadata.CustomReplaces {
		if r.Module == modulePath {
			metadata.CustomReplaces = append(metadata.CustomReplaces[:i], metadata.CustomReplaces[i+1:]...)
			return
		}
	}
}

// createEmpty создает пустые метаданные
func (m *MetadataManager) createEmpty() *types.ProjectMetadata {
	return &types.ProjectMetadata{
		Modules:        make(map[types.ModulePath]types.ModuleMetadata),
		CustomReplaces: make([]types.CustomReplace, 0),
		LastSync:       time.Now(),
	}
}

// LoadOrCreate загружает метаданные или создает новые
func (m *MetadataManager) LoadOrCreate() (*types.ProjectMetadata, error) {
	metadata, err := m.Load()
	if err != nil {
		return nil, err
	}

	if metadata.Modules == nil {
		metadata = m.createEmpty()
	}

	return metadata, nil
}

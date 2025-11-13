package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MaxXxaM/gomm/pkg/logger"
	"github.com/MaxXxaM/gomm/pkg/types"
	"golang.org/x/mod/modfile"
)

// Manager управляет go.work файлом
type Manager struct {
	workspacePath string
	log           *logger.Logger
}

// NewManager создает новый manager
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		workspacePath: "go.work",
		log:           log,
	}
}

// Exists проверяет, существует ли go.work файл
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.workspacePath)
	return err == nil
}

// Read читает существующий go.work файл
func (m *Manager) Read() (*modfile.WorkFile, error) {
	if !m.Exists() {
		return nil, fmt.Errorf("go.work файл не существует")
	}

	data, err := os.ReadFile(m.workspacePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать go.work: %w", err)
	}

	workFile, err := modfile.ParseWork(m.workspacePath, data, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить go.work: %w", err)
	}

	return workFile, nil
}

// GetUsedModules возвращает список модулей в go.work
func (m *Manager) GetUsedModules() ([]string, error) {
	workFile, err := m.Read()
	if err != nil {
		if !m.Exists() {
			return []string{}, nil
		}
		return nil, err
	}

	modules := make([]string, 0, len(workFile.Use))
	for _, use := range workFile.Use {
		modules = append(modules, use.Path)
	}

	return modules, nil
}

// Create создает новый go.work файл с указанными модулями
func (m *Manager) Create(modules []*types.Module, goVersion string) error {
	if goVersion == "" {
		goVersion = "1.21"
	}

	// Создаем новый go.work
	workFile := &modfile.WorkFile{}
	workFile.AddGoStmt(goVersion)

	// Добавляем модули
	for _, module := range modules {
		if module.LocalPath == "" {
			continue
		}

		// Преобразуем путь в относительный
		relPath, err := filepath.Rel(".", module.LocalPath)
		if err != nil {
			m.log.Warning("Не удалось получить относительный путь для %s: %v", module.Path, err)
			relPath = module.LocalPath
		}

		// Очищаем путь
		relPath = filepath.ToSlash(relPath)

		if err := workFile.AddUse(relPath, ""); err != nil {
			return fmt.Errorf("не удалось добавить use директиву для %s: %w", module.Path, err)
		}
	}

	// Форматируем и сохраняем
	data := modfile.Format(workFile.Syntax)
	if err := os.WriteFile(m.workspacePath, data, 0644); err != nil {
		return fmt.Errorf("не удалось сохранить go.work: %w", err)
	}

	m.log.Debug("Создан go.work с %d модулями", len(modules))
	return nil
}

// AddModule добавляет модуль в go.work
func (m *Manager) AddModule(module *types.Module) error {
	if module.LocalPath == "" {
		return fmt.Errorf("модуль %s не имеет локального пути", module.Path)
	}

	var workFile *modfile.WorkFile
	var err error

	if m.Exists() {
		// Читаем существующий файл
		workFile, err = m.Read()
		if err != nil {
			return err
		}
	} else {
		// Создаем новый
		workFile = &modfile.WorkFile{}
		workFile.AddGoStmt("1.21")
	}

	// Преобразуем путь в относительный
	relPath, err := filepath.Rel(".", module.LocalPath)
	if err != nil {
		relPath = module.LocalPath
	}
	relPath = filepath.ToSlash(relPath)

	// Проверяем, не добавлен ли уже модуль
	for _, use := range workFile.Use {
		if use.Path == relPath {
			m.log.Debug("Модуль %s уже в go.work", module.Path)
			return nil
		}
	}

	// Добавляем use директиву
	if err := workFile.AddUse(relPath, ""); err != nil {
		return fmt.Errorf("не удалось добавить use директиву: %w", err)
	}

	// Сохраняем
	data := modfile.Format(workFile.Syntax)
	if err := os.WriteFile(m.workspacePath, data, 0644); err != nil {
		return fmt.Errorf("не удалось сохранить go.work: %w", err)
	}

	m.log.Debug("Добавлен модуль %s в go.work", module.Path)
	return nil
}

// RemoveModule удаляет модуль из go.work
func (m *Manager) RemoveModule(module *types.Module) error {
	if !m.Exists() {
		m.log.Debug("go.work не существует, нечего удалять")
		return nil
	}

	workFile, err := m.Read()
	if err != nil {
		return err
	}

	// Получаем относительный путь
	relPath := module.LocalPath
	if module.LocalPath != "" {
		relPath, _ = filepath.Rel(".", module.LocalPath)
		relPath = filepath.ToSlash(relPath)
	}

	// Ищем и удаляем use директиву
	found := false
	for _, use := range workFile.Use {
		// Проверяем по пути или по module path
		if use.Path == relPath || strings.Contains(use.Path, string(module.Path)) {
			if err := workFile.DropUse(use.Path); err != nil {
				return fmt.Errorf("не удалось удалить use директиву: %w", err)
			}
			found = true
			break
		}
	}

	if !found {
		m.log.Debug("Модуль %s не найден в go.work", module.Path)
		return nil
	}

	// Если больше нет use директив, удаляем файл
	if len(workFile.Use) == 0 {
		m.log.Debug("go.work пуст, удаляем файл")
		return m.Remove()
	}

	// Сохраняем
	data := modfile.Format(workFile.Syntax)
	if err := os.WriteFile(m.workspacePath, data, 0644); err != nil {
		return fmt.Errorf("не удалось сохранить go.work: %w", err)
	}

	m.log.Debug("Удален модуль %s из go.work", module.Path)
	return nil
}

// Remove удаляет go.work файл
func (m *Manager) Remove() error {
	if !m.Exists() {
		return nil
	}

	if err := os.Remove(m.workspacePath); err != nil {
		return fmt.Errorf("не удалось удалить go.work: %w", err)
	}

	m.log.Debug("Удален go.work файл")
	return nil
}

// Update обновляет go.work с новым списком модулей
func (m *Manager) Update(modules []*types.Module) error {
	// Получаем существующую версию Go
	goVersion := "1.21"
	if m.Exists() {
		workFile, err := m.Read()
		if err == nil && workFile.Go != nil {
			goVersion = workFile.Go.Version
		}
	}

	// Создаем новый go.work
	return m.Create(modules, goVersion)
}

// IsModuleInWorkspace проверяет, находится ли модуль в workspace
func (m *Manager) IsModuleInWorkspace(module *types.Module) (bool, error) {
	if !m.Exists() {
		return false, nil
	}

	usedModules, err := m.GetUsedModules()
	if err != nil {
		return false, err
	}

	// Получаем относительный путь модуля
	relPath := module.LocalPath
	if module.LocalPath != "" {
		relPath, _ = filepath.Rel(".", module.LocalPath)
		relPath = filepath.ToSlash(relPath)
	}

	for _, usedPath := range usedModules {
		if usedPath == relPath {
			return true, nil
		}
	}

	return false, nil
}

// GetGoVersion возвращает версию Go из go.work
func (m *Manager) GetGoVersion() (string, error) {
	if !m.Exists() {
		return "", fmt.Errorf("go.work не существует")
	}

	workFile, err := m.Read()
	if err != nil {
		return "", err
	}

	if workFile.Go == nil {
		return "", nil
	}

	return workFile.Go.Version, nil
}

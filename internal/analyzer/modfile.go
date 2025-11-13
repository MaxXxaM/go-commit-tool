package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MaxXxaM/gomm/pkg/types"
	"golang.org/x/mod/modfile"
)

// ModFileParser парсит go.mod файлы
type ModFileParser struct {
	excludePatterns []string
}

// NewModFileParser создает новый парсер
func NewModFileParser(excludePatterns []string) *ModFileParser {
	return &ModFileParser{
		excludePatterns: excludePatterns,
	}
}

// ParseGoMod парсит go.mod файл по указанному пути
func (p *ModFileParser) ParseGoMod(path string) (*types.Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать go.mod: %w", err)
	}

	modFile, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить go.mod: %w", err)
	}

	if modFile.Module == nil {
		return nil, fmt.Errorf("go.mod не содержит объявления module")
	}

	module := &types.Module{
		Path:         types.ModulePath(modFile.Module.Mod.Path),
		Dependencies: make([]*types.Module, 0),
		Dependents:   make([]*types.Module, 0),
	}

	// Парсим версию если есть go.mod в директории git репозитория
	dir := filepath.Dir(path)
	if version := p.getModuleVersion(dir); version != "" {
		module.Version = types.Version(version)
	}

	// Извлекаем зависимости
	for _, req := range modFile.Require {
		if req.Indirect {
			continue // пропускаем indirect зависимости
		}

		// Проверяем исключения
		if p.shouldExclude(req.Mod.Path) {
			continue
		}

		dep := &types.Module{
			Path:       types.ModulePath(req.Mod.Path),
			Version:    types.Version(req.Mod.Version),
			IsExternal: true,
		}

		module.Dependencies = append(module.Dependencies, dep)
	}

	return module, nil
}

// ParseGoModFromDir парсит go.mod из директории
func (p *ModFileParser) ParseGoModFromDir(dir string) (*types.Module, error) {
	goModPath := filepath.Join(dir, "go.mod")
	return p.ParseGoMod(goModPath)
}

// shouldExclude проверяет, должен ли модуль быть исключен
func (p *ModFileParser) shouldExclude(modulePath string) bool {
	for _, pattern := range p.excludePatterns {
		// Простое сопоставление с wildcard (*)
		if matchPattern(pattern, modulePath) {
			return true
		}
	}
	return false
}

// matchPattern проверяет соответствие строки паттерну с wildcard
func matchPattern(pattern, str string) bool {
	// Простая реализация wildcard matching
	if !strings.Contains(pattern, "*") {
		return pattern == str
	}

	parts := strings.Split(pattern, "*")
	if len(parts) == 0 {
		return true
	}

	// Проверяем начало
	if !strings.HasPrefix(str, parts[0]) {
		return false
	}
	str = str[len(parts[0]):]

	// Проверяем середину
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(str, parts[i])
		if idx == -1 {
			return false
		}
		str = str[idx+len(parts[i]):]
	}

	// Проверяем конец
	if len(parts) > 1 {
		return strings.HasSuffix(str, parts[len(parts)-1])
	}

	return true
}

// getModuleVersion пытается получить версию модуля из git тегов
func (p *ModFileParser) getModuleVersion(dir string) string {
	// TODO: реализовать получение версии через git
	// Пока возвращаем пустую строку
	return ""
}

// GetReplaces извлекает replace директивы из go.mod
func (p *ModFileParser) GetReplaces(path string) ([]types.CustomReplace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать go.mod: %w", err)
	}

	modFile, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить go.mod: %w", err)
	}

	replaces := make([]types.CustomReplace, 0, len(modFile.Replace))
	for _, r := range modFile.Replace {
		replaces = append(replaces, types.CustomReplace{
			Module:      types.ModulePath(r.Old.Path),
			Replacement: r.New.Path,
		})
	}

	return replaces, nil
}

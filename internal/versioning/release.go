package versioning

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MaxXxaM/gomm/internal/git"
	"github.com/MaxXxaM/gomm/pkg/logger"
	"github.com/MaxXxaM/gomm/pkg/types"
	"golang.org/x/mod/modfile"
)

// ReleaseManager управляет процессом создания релизов
type ReleaseManager struct {
	gitClient *git.Client
	analyzer  *Analyzer
	config    *types.Config
	log       *logger.Logger
}

// NewReleaseManager создает новый менеджер релизов
func NewReleaseManager(cfg *types.Config, log *logger.Logger) *ReleaseManager {
	return &ReleaseManager{
		gitClient: git.NewClient(log),
		analyzer:  NewAnalyzer(log),
		config:    cfg,
		log:       log,
	}
}

// Release создает релиз для модуля
func (rm *ReleaseManager) Release(module *types.Module, version types.Version, runTests bool) error {
	if module.LocalPath == "" {
		return fmt.Errorf("модуль не имеет локального пути")
	}

	rm.log.Info("Создание релиза %s для %s", version, module.Path)

	// Проверяем, что это git репозиторий
	if !rm.gitClient.IsRepo(module.LocalPath) {
		return fmt.Errorf("директория не является git репозиторием: %s", module.LocalPath)
	}

	// Проверяем наличие незакоммиченных изменений
	hasChanges, err := rm.gitClient.HasUncommittedChanges(module.LocalPath)
	if err != nil {
		return fmt.Errorf("не удалось проверить изменения: %w", err)
	}

	if hasChanges {
		return fmt.Errorf("есть незакоммиченные изменения, закоммитьте их перед релизом")
	}

	// Запускаем тесты если требуется
	if runTests {
		rm.log.Info("Запуск тестов...")
		if err := rm.runTests(module.LocalPath); err != nil {
			return fmt.Errorf("тесты не прошли: %w", err)
		}
		rm.log.Success("Тесты пройдены")
	}

	// Валидируем версию
	if err := ValidateVersion(version); err != nil {
		return fmt.Errorf("невалидная версия: %w", err)
	}

	// Создаем тег
	if err := rm.createTag(module.LocalPath, string(version)); err != nil {
		return fmt.Errorf("не удалось создать тег: %w", err)
	}

	rm.log.Success("Тег %s создан", version)

	// Push тега если настроено
	if rm.config.Release.AutoPush {
		if err := rm.pushTag(module.LocalPath, string(version)); err != nil {
			return fmt.Errorf("не удалось запушить тег: %w", err)
		}
		rm.log.Success("Тег %s запушен", version)
	}

	// Обновляем версию модуля
	module.Version = version

	return nil
}

// createTag создает git тег
func (rm *ReleaseManager) createTag(dir, tag string) error {
	cmd := exec.Command("git", "-C", dir, "tag", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка создания тега: %w\nВывод: %s", err, string(output))
	}
	return nil
}

// pushTag пушит тег в remote
func (rm *ReleaseManager) pushTag(dir, tag string) error {
	cmd := exec.Command("git", "-C", dir, "push", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка push тега: %w\nВывод: %s", err, string(output))
	}
	return nil
}

// runTests запускает тесты для модуля
func (rm *ReleaseManager) runTests(dir string) error {
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("тесты не прошли: %w\nВывод: %s", err, string(output))
	}
	return nil
}

// UpdateDependency обновляет зависимость в go.mod
func (rm *ReleaseManager) UpdateDependency(modulePath string, depPath types.ModulePath, newVersion types.Version) error {
	goModPath := filepath.Join(modulePath, "go.mod")

	// Читаем go.mod
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать go.mod: %w", err)
	}

	modFile, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return fmt.Errorf("не удалось распарсить go.mod: %w", err)
	}

	// Ищем зависимость
	found := false
	for _, req := range modFile.Require {
		if req.Mod.Path == string(depPath) {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("зависимость %s не найдена в go.mod", depPath)
	}

	// Удаляем старую require
	if err := modFile.DropRequire(string(depPath)); err != nil {
		return fmt.Errorf("не удалось удалить старую require: %w", err)
	}

	// Добавляем новую require с новой версией
	if err := modFile.AddRequire(string(depPath), string(newVersion)); err != nil {
		return fmt.Errorf("не удалось добавить новую require: %w", err)
	}

	// Сохраняем go.mod
	formatted := modfile.Format(modFile.Syntax)
	if err := os.WriteFile(goModPath, formatted, 0644); err != nil {
		return fmt.Errorf("не удалось сохранить go.mod: %w", err)
	}

	rm.log.Debug("Обновлена зависимость %s до %s в %s", depPath, newVersion, modulePath)

	// Запускаем go mod tidy
	if err := rm.runGoModTidy(modulePath); err != nil {
		return fmt.Errorf("не удалось выполнить go mod tidy: %w", err)
	}

	return nil
}

// runGoModTidy запускает go mod tidy
func (rm *ReleaseManager) runGoModTidy(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка go mod tidy: %w\nВывод: %s", err, string(output))
	}
	return nil
}

// CommitChanges коммитит изменения
func (rm *ReleaseManager) CommitChanges(dir, message string) error {
	// Добавляем все изменения
	cmd := exec.Command("git", "-C", dir, "add", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("не удалось добавить файлы: %w\nВывод: %s", err, string(output))
	}

	// Коммитим
	cmd = exec.Command("git", "-C", dir, "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Проверяем, может нет изменений для коммита
		if strings.Contains(string(output), "nothing to commit") {
			rm.log.Debug("Нет изменений для коммита")
			return nil
		}
		return fmt.Errorf("не удалось создать коммит: %w\nВывод: %s", err, string(output))
	}

	return nil
}

// PushChanges пушит изменения
func (rm *ReleaseManager) PushChanges(dir string) error {
	cmd := exec.Command("git", "-C", dir, "push")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("не удалось запушить изменения: %w\nВывод: %s", err, string(output))
	}
	return nil
}

// CascadeRelease создает релизы для всех модулей в правильном порядке
func (rm *ReleaseManager) CascadeRelease(tree *types.DependencyTree, modules []*types.Module, bump types.VersionBump) error {
	rm.log.Info("Каскадный релиз для %d модулей", len(modules))

	// Используем топологический порядок из дерева
	// Релизим от листьев к корню
	releasedModules := make(map[types.ModulePath]types.Version)

	for _, module := range tree.TopologicalOrder {
		// Проверяем, входит ли модуль в список для релиза
		shouldRelease := false
		var targetModule *types.Module
		for _, m := range modules {
			if m.Path == module.Path {
				shouldRelease = true
				targetModule = m
				break
			}
		}

		if !shouldRelease {
			continue
		}

		rm.log.Info("Релиз модуля: %s", module.Path)

		// Получаем текущую версию
		currentVersion, err := rm.analyzer.GetCurrentVersion(targetModule.LocalPath)
		if err != nil {
			return fmt.Errorf("не удалось получить версию для %s: %w", module.Path, err)
		}

		// Вычисляем новую версию
		newVersion, err := BumpVersion(currentVersion, bump)
		if err != nil {
			return fmt.Errorf("не удалось вычислить новую версию для %s: %w", module.Path, err)
		}

		// Обновляем зависимости в этом модуле если они были релизнуты
		for depPath, depVersion := range releasedModules {
			// Проверяем, зависит ли текущий модуль от уже релизнутого
			for _, dep := range module.Dependencies {
				if dep.Path == depPath {
					rm.log.Info("Обновление зависимости %s до %s", depPath, depVersion)
					if err := rm.UpdateDependency(targetModule.LocalPath, depPath, depVersion); err != nil {
						return fmt.Errorf("не удалось обновить зависимость: %w", err)
					}
				}
			}
		}

		// Коммитим изменения если есть
		commitMsg := fmt.Sprintf("Update dependencies before release %s", newVersion)
		if err := rm.CommitChanges(targetModule.LocalPath, commitMsg); err != nil {
			return fmt.Errorf("не удалось закоммитить изменения: %w", err)
		}

		// Создаем релиз
		if err := rm.Release(targetModule, newVersion, rm.config.Release.RunTests); err != nil {
			return fmt.Errorf("не удалось создать релиз для %s: %w", module.Path, err)
		}

		// Сохраняем информацию о релизе
		releasedModules[module.Path] = newVersion
		rm.log.Success("Релиз %s создан для %s", newVersion, module.Path)
	}

	return nil
}

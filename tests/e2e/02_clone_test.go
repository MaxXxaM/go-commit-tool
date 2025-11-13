package e2e

import (
	"path/filepath"
	"testing"

	"github.com/MaxXxaM/gomm/tests/helpers"
)

// TestCloneSingleDependency тестирует клонирование одной зависимости
func TestCloneSingleDependency(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем зависимость
	depModulePath := "github.com/test/dep-a"
	depDir := env.CreateTestModule(depModulePath)

	env.CreateTestProject("github.com/test/myproject", depModulePath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем зависимость
	output := env.RunGommExpectSuccess("clone", depModulePath)

	// Проверяем, что зависимость склонирована
	env.AssertContains(output, "cloned")
	env.AssertContains(output, depModulePath)

	// Проверяем, что директория создана (в local или shared)
	localPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")
	sharedPath := filepath.Join(env.SharedDir, "dep-a")

	if !env.DirExists(localPath) && !env.DirExists(sharedPath) {
		t.Fatal("Expected dependency to be cloned to either local or shared directory")
	}

	// Проверяем, что метаданные обновлены
	metadata := env.LoadMetadata()
	if metadata["modules"] == nil {
		t.Fatal("Expected modules in metadata")
	}
}

// TestCloneWithLocalMode тестирует клонирование в local режиме
func TestCloneWithLocalMode(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем в local режиме
	output := env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	env.AssertContains(output, "local")

	// Проверяем, что зависимость в local директории (рядом с проектом)
	localPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")
	env.AssertDirExists(localPath)

	// Проверяем, что это git репозиторий
	gitDir := filepath.Join(localPath, ".git")
	env.AssertDirExists(gitDir)
}

// TestCloneWithSharedMode тестирует клонирование в shared режиме
func TestCloneWithSharedMode(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем в shared режиме
	output := env.RunGommExpectSuccess("clone", depModulePath, "--mode", "shared")

	env.AssertContains(output, "shared")

	// Проверяем, что зависимость в shared директории
	sharedPath := filepath.Join(env.SharedDir, "dep-a")
	env.AssertDirExists(sharedPath)
}

// TestCloneAll тестирует клонирование всех зависимостей
func TestCloneAll(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем несколько зависимостей
	depAPath := "github.com/test/dep-a"
	depBPath := "github.com/test/dep-b"

	env.CreateTestModule(depAPath)
	env.CreateTestModule(depBPath)

	env.CreateTestProject("github.com/test/myproject", depAPath, depBPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем все зависимости
	output := env.RunGommExpectSuccess("clone", "--all")

	// Проверяем, что обе зависимости склонированы
	env.AssertContains(output, depAPath)
	env.AssertContains(output, depBPath)
}

// TestCloneAlreadyCloned тестирует повторное клонирование
func TestCloneAlreadyCloned(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Первое клонирование
	env.RunGommExpectSuccess("clone", depModulePath)

	// Второе клонирование (должно быть пропущено или обновлено)
	output := env.RunGommExpectSuccess("clone", depModulePath)

	// Проверяем, что есть сообщение о том, что уже склонировано
	env.AssertContains(output, "already")
}

// TestMoveLocalToShared тестирует перемещение модуля из local в shared
func TestMoveLocalToShared(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем в local
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	localPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")
	env.AssertDirExists(localPath)

	// Перемещаем в shared
	output := env.RunGommExpectSuccess("move", depModulePath, "--to", "shared")

	env.AssertContains(output, "moved")

	// Проверяем, что модуль теперь в shared
	sharedPath := filepath.Join(env.SharedDir, "dep-a")
	env.AssertDirExists(sharedPath)

	// Проверяем, что local директория удалена
	env.AssertFileNotExists(localPath)
}

// TestMoveSharedToLocal тестирует перемещение модуля из shared в local
func TestMoveSharedToLocal(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем в shared
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "shared")

	sharedPath := filepath.Join(env.SharedDir, "dep-a")
	env.AssertDirExists(sharedPath)

	// Перемещаем в local
	output := env.RunGommExpectSuccess("move", depModulePath, "--to", "local")

	env.AssertContains(output, "moved")

	// Проверяем, что модуль теперь в local
	localPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")
	env.AssertDirExists(localPath)

	// Проверяем, что shared директория удалена
	env.AssertFileNotExists(sharedPath)
}

// TestCleanSingleDependency тестирует удаление одной зависимости
func TestCleanSingleDependency(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	localPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")
	env.AssertDirExists(localPath)

	// Удаляем зависимость
	output := env.RunGommExpectSuccess("clean", depModulePath)

	env.AssertContains(output, "cleaned")

	// Проверяем, что директория удалена
	env.AssertFileNotExists(localPath)
}

// TestCleanWithUncommittedChanges тестирует удаление с несохраненными изменениями
func TestCleanWithUncommittedChanges(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	localPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Вносим изменения
	env.WriteFile(filepath.Join(localPath, "test.txt"), "test content")

	// Пытаемся удалить без --force (должна быть ошибка)
	output := env.RunGommExpectError("clean", depModulePath)

	env.AssertContains(output, "uncommitted")

	// Удаляем с --force
	output = env.RunGommExpectSuccess("clean", depModulePath, "--force")

	env.AssertContains(output, "cleaned")
	env.AssertFileNotExists(localPath)
}

// TestCleanAll тестирует удаление всех зависимостей
func TestCleanAll(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depAPath := "github.com/test/dep-a"
	depBPath := "github.com/test/dep-b"

	env.CreateTestModule(depAPath)
	env.CreateTestModule(depBPath)

	env.CreateTestProject("github.com/test/myproject", depAPath, depBPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", "--all", "--mode", "local")

	// Удаляем все зависимости
	output := env.RunGommExpectSuccess("clean", "--all", "--force")

	// Проверяем, что обе удалены
	env.AssertContains(output, depAPath)
	env.AssertContains(output, depBPath)
}

// TestCloneNonExistentDependency тестирует клонирование несуществующей зависимости
func TestCloneNonExistentDependency(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")
	env.RunGommExpectSuccess("init")

	// Пытаемся клонировать несуществующую зависимость
	output := env.RunGommExpectError("clone", "github.com/nonexistent/module")

	// Должна быть ошибка
	env.AssertContains(output, "not found")
}

// TestPlacementRules тестирует автоматическое определение placement
func TestPlacementRules(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем модули, соответствующие разным паттернам
	commonModulePath := "github.com/test/common-lib"
	serviceModulePath := "github.com/test/service-api"

	env.CreateTestModule(commonModulePath)
	env.CreateTestModule(serviceModulePath)

	env.CreateTestProject("github.com/test/myproject", commonModulePath, serviceModulePath)

	env.RunGommExpectSuccess("init")

	// Настраиваем правила placement в конфиге
	configPath := filepath.Join(env.ProjectDir, ".gomm.yaml")
	cfg := env.LoadGommConfig()

	// Примечание: это зависит от реализации Config структуры
	// Может потребоваться ручное редактирование YAML
	configContent := `shared_dir: ` + env.SharedDir + `
default_placement: auto

placement_rules:
  - pattern: "github.com/test/common-*"
    mode: shared
  - pattern: "github.com/test/service-*"
    mode: local

versioning:
  auto_detect: true

release:
  auto_push: false
  create_changelog: false
  run_tests: true
`
	env.WriteFile(configPath, configContent)

	env.RunGommExpectSuccess("scan")

	// Клонируем модули с auto режимом
	env.RunGommExpectSuccess("clone", commonModulePath, "--mode", "auto")
	env.RunGommExpectSuccess("clone", serviceModulePath, "--mode", "auto")

	// Проверяем, что common-lib в shared
	sharedPath := filepath.Join(env.SharedDir, "common-lib")
	env.AssertDirExists(sharedPath)

	// Проверяем, что service-api в local
	localPath := filepath.Join(filepath.Dir(env.ProjectDir), "service-api")
	env.AssertDirExists(localPath)
}

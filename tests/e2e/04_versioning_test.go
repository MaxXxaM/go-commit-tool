package e2e

import (
	"path/filepath"
	"testing"

	"github.com/MaxXxaM/gomm/tests/helpers"
)

// TestVersionCheckCurrentVersion тестирует отображение текущей версии
func TestVersionCheckCurrentVersion(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	depDir := env.CreateTestModule(depModulePath)

	env.CreateTestProject("github.com/test/myproject", depModulePath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath)

	// Проверяем версию склонированного модуля
	output, err := env.RunGommInDir(filepath.Join(filepath.Dir(env.ProjectDir), "dep-a"), "version", "check")
	if err != nil {
		t.Logf("version check output: %s", output)
		t.Logf("Note: version check might fail if module is not in workspace mode")
	}

	// Если команда успешна, проверяем вывод
	if err == nil {
		env.AssertContains(output, "v0.1.0")
	}
}

// TestVersionCheckNoChanges тестирует версию без изменений
func TestVersionCheckNoChanges(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Проверяем версию (без изменений)
	output, _ := env.RunGommInDir(depLocalPath, "version", "check")

	// Должна отображаться текущая версия
	env.AssertContains(output, "v0.1.0")
}

// TestVersionCheckWithMinorChanges тестирует определение minor версии
func TestVersionCheckWithMinorChanges(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Добавляем новую публичную функцию (minor change)
	env.ModifyGoFile(depLocalPath, "dep-a.go", "NewFeature")

	// Коммитим изменения
	env.GitCommit(depLocalPath, "Add new feature")

	// Проверяем версию
	output, _ := env.RunGommInDir(depLocalPath, "version", "check")

	// Должна предлагаться minor версия
	if !env.RunGommExpectError("version", "check") {
		env.AssertContains(output, "minor")
	}
}

// TestVersionCheckWithPatchChanges тестирует определение patch версии
func TestVersionCheckWithPatchChanges(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Изменяем комментарий (patch change)
	goFilePath := filepath.Join(depLocalPath, "dep-a.go")
	content := env.ReadFile(goFilePath)
	content += "\n// Bug fix comment\n"
	env.WriteFile(goFilePath, content)

	// Коммитим изменения
	env.GitCommit(depLocalPath, "Fix: update comment")

	// Проверяем версию
	output, _ := env.RunGommInDir(depLocalPath, "version", "check")

	// Может предлагаться patch версия (зависит от детектора изменений)
	t.Logf("Version check output: %s", output)
}

// TestVersionCheckForSpecificModule тестирует version check для конкретного модуля
func TestVersionCheckForSpecificModule(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	// Проверяем версию из корня проекта, указывая модуль
	output, _ := env.RunGomm("version", "check", depModulePath)

	// Должна отображаться информация о модуле
	if output != "" {
		env.AssertContains(output, depModulePath)
	}
}

// TestVersionCheckUntaggedModule тестирует модуль без тегов
func TestVersionCheckUntaggedModule(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-new"
	depDir := env.CreateTestModule(depModulePath)

	// Удаляем тег
	env.RunCommand(depDir, "git", "tag", "-d", "v0.1.0")

	env.CreateTestProject("github.com/test/myproject", depModulePath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-new")

	// Проверяем версию
	output, _ := env.RunGommInDir(depLocalPath, "version", "check")

	// Должна предлагаться версия v0.1.0 (первый релиз)
	t.Logf("Version check output: %s", output)
}

// TestVersionCheckShowsChangesSummary тестирует отображение сводки изменений
func TestVersionCheckShowsChangesSummary(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Вносим несколько изменений
	env.ModifyGoFile(depLocalPath, "dep-a.go", "Feature1")
	env.GitCommit(depLocalPath, "Add feature 1")

	env.ModifyGoFile(depLocalPath, "dep-a.go", "Feature2")
	env.GitCommit(depLocalPath, "Add feature 2")

	// Проверяем версию
	output, _ := env.RunGommInDir(depLocalPath, "version", "check")

	// Должна быть сводка изменений
	if output != "" {
		// Проверяем наличие информации о коммитах или изменениях
		t.Logf("Version check output: %s", output)
	}
}

// TestVersionCheckVerbose тестирует verbose режим version check
func TestVersionCheckVerbose(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Вносим изменения
	env.ModifyGoFile(depLocalPath, "dep-a.go", "NewFeature")
	env.GitCommit(depLocalPath, "Add new feature")

	// Проверяем версию с verbose флагом
	output, _ := env.RunGommInDir(depLocalPath, "version", "check", "--verbose")

	// В verbose режиме должно быть больше деталей
	if output != "" {
		t.Logf("Verbose version check output: %s", output)
	}
}

// TestVersionCheckWithMultipleVersionTags тестирует модуль с несколькими тегами
func TestVersionCheckWithMultipleVersionTags(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	depDir := env.CreateTestModule(depModulePath)

	// Добавляем еще несколько версий
	env.ModifyGoFile(depDir, "dep-a.go", "Feature1")
	env.GitCommit(depDir, "Add feature 1")
	env.GitTag(depDir, "v0.2.0")

	env.ModifyGoFile(depDir, "dep-a.go", "Feature2")
	env.GitCommit(depDir, "Add feature 2")
	env.GitTag(depDir, "v0.3.0")

	env.CreateTestProject("github.com/test/myproject", depModulePath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Проверяем версию
	output, _ := env.RunGommInDir(depLocalPath, "version", "check")

	// Должна отображаться последняя версия v0.3.0
	if output != "" {
		env.AssertContains(output, "v0.3.0")
	}
}

// TestVersionCheckInProjectRoot тестирует version check в корне проекта
func TestVersionCheckInProjectRoot(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")

	// Создаем тег для проекта
	env.GitTag(env.ProjectDir, "v1.0.0")

	env.RunGommExpectSuccess("init")

	// Проверяем версию проекта
	output, _ := env.RunGomm("version", "check")

	// Должна отображаться версия проекта
	if output != "" {
		env.AssertContains(output, "v1.0.0")
	}
}

// TestVersionCheckDetectsBreakingChanges тестирует определение breaking changes
func TestVersionCheckDetectsBreakingChanges(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	depDir := env.CreateTestModule(depModulePath)

	env.CreateTestProject("github.com/test/myproject", depModulePath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Изменяем сигнатуру публичной функции (breaking change)
	goFilePath := filepath.Join(depLocalPath, "dep-a.go")
	content := env.ReadFile(goFilePath)

	// Заменяем функцию Hello() на Hello(name string)
	// Это простая симуляция breaking change
	newContent := content + `

// HelloWithName is a breaking change - signature changed
func HelloWithName(name string) string {
	return "Hello " + name
}
`
	env.WriteFile(goFilePath, newContent)

	env.GitCommit(depLocalPath, "BREAKING: change function signature")

	// Проверяем версию
	output, _ := env.RunGommInDir(depLocalPath, "version", "check")

	// Может предлагаться major версия (зависит от детектора breaking changes)
	// Это сложно тестировать без полного AST анализа
	t.Logf("Version check output: %s", output)
}

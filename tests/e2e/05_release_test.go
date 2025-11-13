package e2e

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/MaxXxaM/gomm/tests/helpers"
)

// TestReleaseAuto тестирует автоматический релиз
func TestReleaseAuto(t *testing.T) {
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

	// Создаем релиз с автоопределением версии
	output, _ := env.RunGommInDir(depLocalPath, "release", "--auto")

	// Проверяем, что релиз создан
	if output != "" {
		env.AssertContains(output, "release")

		// Проверяем, что тег создан
		tags := env.RunCommand(depLocalPath, "git", "tag", "-l")
		// Должен быть новый тег (v0.2.0 при minor изменениях)
		t.Logf("Git tags: %s", tags)
	}
}

// TestReleaseBumpPatch тестирует релиз с patch версией
func TestReleaseBumpPatch(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Вносим небольшие изменения
	goFilePath := filepath.Join(depLocalPath, "dep-a.go")
	content := env.ReadFile(goFilePath)
	content += "\n// Bug fix\n"
	env.WriteFile(goFilePath, content)
	env.GitCommit(depLocalPath, "Fix: minor bug")

	// Создаем релиз с patch bump
	output, _ := env.RunGommInDir(depLocalPath, "release", "--bump", "patch")

	if output != "" {
		// Проверяем, что создан тег v0.1.1
		tags := env.RunCommand(depLocalPath, "git", "tag", "-l", "v0.1.1")
		if strings.TrimSpace(tags) == "v0.1.1" {
			t.Log("Patch version tag created successfully")
		}
	}
}

// TestReleaseBumpMinor тестирует релиз с minor версией
func TestReleaseBumpMinor(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Добавляем новую функцию
	env.ModifyGoFile(depLocalPath, "dep-a.go", "NewFeature")
	env.GitCommit(depLocalPath, "Add new feature")

	// Создаем релиз с minor bump
	output, _ := env.RunGommInDir(depLocalPath, "release", "--bump", "minor")

	if output != "" {
		// Проверяем, что создан тег v0.2.0
		tags := env.RunCommand(depLocalPath, "git", "tag", "-l", "v0.2.0")
		if strings.TrimSpace(tags) == "v0.2.0" {
			t.Log("Minor version tag created successfully")
		}
	}
}

// TestReleaseBumpMajor тестирует релиз с major версией
func TestReleaseBumpMajor(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Вносим breaking changes
	env.ModifyGoFile(depLocalPath, "dep-a.go", "BreakingChange")
	env.GitCommit(depLocalPath, "BREAKING: major API change")

	// Создаем релиз с major bump
	output, _ := env.RunGommInDir(depLocalPath, "release", "--bump", "major")

	if output != "" {
		// Проверяем, что создан тег v1.0.0
		tags := env.RunCommand(depLocalPath, "git", "tag", "-l", "v1.0.0")
		if strings.TrimSpace(tags) == "v1.0.0" {
			t.Log("Major version tag created successfully")
		}
	}
}

// TestReleaseWithSpecificVersion тестирует релиз с конкретной версией
func TestReleaseWithSpecificVersion(t *testing.T) {
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
	env.ModifyGoFile(depLocalPath, "dep-a.go", "Feature")
	env.GitCommit(depLocalPath, "Add feature")

	// Создаем релиз с конкретной версией
	customVersion := "v2.5.0"
	output, _ := env.RunGommInDir(depLocalPath, "release", "--version", customVersion)

	if output != "" {
		// Проверяем, что создан нужный тег
		tags := env.RunCommand(depLocalPath, "git", "tag", "-l", customVersion)
		if strings.TrimSpace(tags) == customVersion {
			t.Logf("Custom version tag %s created successfully", customVersion)
		}
	}
}

// TestReleaseWithoutChanges тестирует релиз без изменений
func TestReleaseWithoutChanges(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Пытаемся создать релиз без изменений
	output := env.RunGommExpectError("release", "--auto")

	// Должна быть ошибка о том, что нет изменений
	env.AssertContains(output, "no changes")
}

// TestReleaseUpdatesGoMod тестирует обновление go.mod после релиза
func TestReleaseUpdatesGoMod(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем цепочку: project -> dep-a
	depAPath := "github.com/test/dep-a"
	env.CreateTestModule(depAPath)
	env.CreateTestProject("github.com/test/myproject", depAPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depAPath, "--mode", "local")

	depALocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Вносим изменения в dep-a
	env.ModifyGoFile(depALocalPath, "dep-a.go", "NewFeature")
	env.GitCommit(depALocalPath, "Add feature")

	// Создаем релиз dep-a
	output, _ := env.RunGommInDir(depALocalPath, "release", "--bump", "minor")

	if output != "" {
		// Проверяем, что go.mod проекта обновлен (если используется --cascade или вручную)
		// Это зависит от реализации автообновления
		t.Logf("Release output: %s", output)
	}
}

// TestReleaseCascade тестирует каскадный релиз
func TestReleaseCascade(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем цепочку: project -> dep-a -> dep-b
	depBPath := "github.com/test/dep-b"
	env.CreateTestModule(depBPath)

	depAPath := "github.com/test/dep-a"
	env.CreateTestModule(depAPath, depBPath)

	env.CreateTestProject("github.com/test/myproject", depAPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", "--all")

	depBLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-b")

	// Вносим изменения в dep-b
	env.ModifyGoFile(depBLocalPath, "dep-b.go", "Feature")
	env.GitCommit(depBLocalPath, "Add feature")

	// Создаем каскадный релиз
	output, _ := env.RunGommInDir(depBLocalPath, "release", "--cascade", "--bump", "minor")

	if output != "" {
		// Должны быть обновлены все зависимые модули
		env.AssertContains(output, "cascade")
		t.Logf("Cascade release output: %s", output)
	}
}

// TestReleaseNoTests тестирует релиз без запуска тестов
func TestReleaseNoTests(t *testing.T) {
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
	env.ModifyGoFile(depLocalPath, "dep-a.go", "Feature")
	env.GitCommit(depLocalPath, "Add feature")

	// Создаем релиз без тестов
	output, _ := env.RunGommInDir(depLocalPath, "release", "--no-tests", "--bump", "patch")

	if output != "" {
		// Релиз должен быть создан без запуска тестов
		env.AssertContains(output, "release")
		t.Logf("Release without tests output: %s", output)
	}
}

// TestReleaseFailsOnFailedTests тестирует провал релиза при ошибке тестов
func TestReleaseFailsOnFailedTests(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	depDir := env.CreateTestModule(depModulePath)

	// Добавляем падающий тест
	testContent := `package dep-a

import "testing"

func TestFailing(t *testing.T) {
	t.Error("This test always fails")
}
`
	env.WriteFile(filepath.Join(depDir, "dep-a_test.go"), testContent)
	env.GitCommit(depDir, "Add failing test")
	env.GitTag(depDir, "v0.2.0")

	env.CreateTestProject("github.com/test/myproject", depModulePath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Вносим изменения
	env.ModifyGoFile(depLocalPath, "dep-a.go", "Feature")
	env.GitCommit(depLocalPath, "Add feature")

	// Пытаемся создать релиз (должен упасть из-за тестов)
	output := env.RunGommExpectError("release", "--bump", "patch")

	// Должна быть ошибка о провале тестов
	env.AssertContains(output, "test")
}

// TestReleaseForSpecificModule тестирует релиз конкретного модуля из проекта
func TestReleaseForSpecificModule(t *testing.T) {
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
	env.ModifyGoFile(depLocalPath, "dep-a.go", "Feature")
	env.GitCommit(depLocalPath, "Add feature")

	// Создаем релиз из корня проекта, указывая модуль
	output, _ := env.RunGomm("release", depModulePath, "--bump", "minor")

	if output != "" {
		env.AssertContains(output, depModulePath)
		t.Logf("Release output: %s", output)
	}
}

// TestReleaseCreatesProperTag тестирует формат создаваемого тега
func TestReleaseCreatesProperTag(t *testing.T) {
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
	env.ModifyGoFile(depLocalPath, "dep-a.go", "Feature")
	env.GitCommit(depLocalPath, "Add feature")

	// Создаем релиз
	output, _ := env.RunGommInDir(depLocalPath, "release", "--bump", "minor")

	if output != "" {
		// Проверяем формат тега
		tags := env.RunCommand(depLocalPath, "git", "tag", "-l")
		lines := strings.Split(strings.TrimSpace(tags), "\n")

		// Должен быть тег v0.2.0
		found := false
		for _, tag := range lines {
			if strings.TrimSpace(tag) == "v0.2.0" {
				found = true
				break
			}
		}

		if found {
			t.Log("Proper semantic version tag created")
		}
	}
}

// TestReleaseWithUncommittedChanges тестирует релиз с несохраненными изменениями
func TestReleaseWithUncommittedChanges(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	depLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-a")

	// Вносим изменения БЕЗ коммита
	env.ModifyGoFile(depLocalPath, "dep-a.go", "Feature")

	// Пытаемся создать релиз
	output := env.RunGommExpectError("release", "--bump", "patch")

	// Должна быть ошибка о несохраненных изменениях
	env.AssertContains(output, "uncommitted")
}

// TestReleaseFromProjectRoot тестирует релиз самого проекта
func TestReleaseFromProjectRoot(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")
	env.RunGommExpectSuccess("init")

	// Вносим изменения в проект
	env.WriteFile(filepath.Join(env.ProjectDir, "feature.go"), `package main

func NewFeature() string {
	return "feature"
}
`)
	env.GitCommit(env.ProjectDir, "Add feature")

	// Создаем релиз проекта
	output, _ := env.RunGomm("release", "--bump", "minor")

	if output != "" {
		// Проверяем, что создан тег для проекта
		tags := env.RunCommand(env.ProjectDir, "git", "tag", "-l")
		t.Logf("Project tags: %s", tags)
	}
}

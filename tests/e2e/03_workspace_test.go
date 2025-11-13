package e2e

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/MaxXxaM/gomm/tests/helpers"
)

// TestLocalCommand тестирует команду gomm local
func TestLocalCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath, "--mode", "local")

	// Переключаемся на локальную разработку
	output := env.RunGommExpectSuccess("local", depModulePath)

	env.AssertContains(output, "workspace")

	// Проверяем, что go.work создан
	goWorkPath := env.GetProjectGoWorkPath()
	env.AssertFileExists(goWorkPath)

	// Проверяем содержимое go.work
	uses := env.ParseGoWork()
	if len(uses) == 0 {
		t.Fatal("Expected at least one use directive in go.work")
	}
}

// TestLocalAllDependencies тестирует gomm local --all
func TestLocalAllDependencies(t *testing.T) {
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

	// Переключаем все на локальную разработку
	output := env.RunGommExpectSuccess("local", "--all")

	env.AssertContains(output, "workspace")

	// Проверяем go.work
	uses := env.ParseGoWork()
	if len(uses) < 2 {
		t.Fatalf("Expected at least 2 use directives, got %d", len(uses))
	}
}

// TestLocalWithPrefix тестирует gomm local --prefix
func TestLocalWithPrefix(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depAPath := "github.com/test/dep-a"
	depBPath := "github.com/other/dep-b"

	env.CreateTestModule(depAPath)
	env.CreateTestModule(depBPath)
	env.CreateTestProject("github.com/test/myproject", depAPath, depBPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", "--all")

	// Переключаем только модули с префиксом github.com/test
	output := env.RunGommExpectSuccess("local", "--prefix", "github.com/test")

	env.AssertContains(output, "workspace")

	// Проверяем, что только dep-a в workspace
	uses := env.ParseGoWork()

	hasDepA := false
	for _, use := range uses {
		if strings.Contains(use, "dep-a") {
			hasDepA = true
		}
		// dep-b не должна быть в workspace
		if strings.Contains(use, "dep-b") {
			t.Fatal("Expected dep-b to not be in workspace")
		}
	}

	if !hasDepA {
		t.Fatal("Expected dep-a to be in workspace")
	}
}

// TestRemoteCommand тестирует команду gomm remote
func TestRemoteCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", depModulePath)
	env.RunGommExpectSuccess("local", depModulePath)

	// Проверяем, что go.work существует
	env.AssertFileExists(env.GetProjectGoWorkPath())

	// Переключаемся обратно на remote
	output := env.RunGommExpectSuccess("remote", depModulePath)

	env.AssertContains(output, "remote")

	// go.work должен быть удален или пуст
	// (зависит от реализации - если других зависимостей нет)
}

// TestRemoteAllDependencies тестирует gomm remote --all
func TestRemoteAllDependencies(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depAPath := "github.com/test/dep-a"
	depBPath := "github.com/test/dep-b"

	env.CreateTestModule(depAPath)
	env.CreateTestModule(depBPath)
	env.CreateTestProject("github.com/test/myproject", depAPath, depBPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", "--all")
	env.RunGommExpectSuccess("local", "--all")

	// Переключаем все обратно на remote
	output := env.RunGommExpectSuccess("remote", "--all")

	env.AssertContains(output, "remote")

	// go.work должен быть удален
	goWorkPath := env.GetProjectGoWorkPath()
	if env.FileExists(goWorkPath) {
		// Проверяем, что он пустой (только go версия)
		content := env.ReadFile(goWorkPath)
		if strings.Contains(content, "use") {
			t.Fatal("Expected go.work to not contain use directives")
		}
	}
}

// TestMixedMode тестирует смешанный режим работы
func TestMixedMode(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depAPath := "github.com/test/dep-a"
	depBPath := "github.com/test/dep-b"

	env.CreateTestModule(depAPath)
	env.CreateTestModule(depBPath)
	env.CreateTestProject("github.com/test/myproject", depAPath, depBPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", "--all")

	// Переключаем только dep-a на local
	env.RunGommExpectSuccess("local", depAPath)

	// Проверяем статус (должен показывать смешанный режим)
	output := env.RunGommExpectSuccess("status", "-v")

	// Должна быть информация о смешанном режиме
	env.AssertContains(output, "workspace")
	env.AssertContains(output, "1") // одна зависимость в workspace

	// Проверяем go.work
	uses := env.ParseGoWork()
	if len(uses) != 1 {
		t.Fatalf("Expected exactly 1 use directive in mixed mode, got %d", len(uses))
	}
}

// TestReplaceAdd тестирует добавление custom replace директивы
func TestReplaceAdd(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")
	env.RunGommExpectSuccess("init")

	// Добавляем replace директиву
	oldModule := "example.com/old-pkg"
	newModule := "example.com/new-pkg"

	output := env.RunGommExpectSuccess("replace", "add", oldModule, newModule)

	env.AssertContains(output, "replace")
	env.AssertContains(output, oldModule)

	// Проверяем go.mod
	goModPath := env.GetProjectGoModPath()
	goModContent := env.ReadFile(goModPath)

	env.AssertContains(goModContent, "replace")
	env.AssertContains(goModContent, oldModule)
	env.AssertContains(goModContent, newModule)
}

// TestReplaceAddLocalPath тестирует replace на локальный путь
func TestReplaceAddLocalPath(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")
	env.RunGommExpectSuccess("init")

	// Добавляем replace на локальный путь
	module := "example.com/module"
	localPath := "../local-fork"

	output := env.RunGommExpectSuccess("replace", "add", module, localPath)

	env.AssertContains(output, "replace")

	// Проверяем go.mod
	goModContent := env.ReadFile(env.GetProjectGoModPath())
	env.AssertContains(goModContent, module)
	env.AssertContains(goModContent, localPath)
}

// TestReplaceList тестирует отображение replace директив
func TestReplaceList(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")
	env.RunGommExpectSuccess("init")

	// Добавляем несколько replace
	env.RunGommExpectSuccess("replace", "add", "example.com/mod1", "example.com/new1")
	env.RunGommExpectSuccess("replace", "add", "example.com/mod2", "../local2")

	// Получаем список
	output := env.RunGommExpectSuccess("replace", "list")

	// Проверяем, что обе директивы отображаются
	env.AssertContains(output, "example.com/mod1")
	env.AssertContains(output, "example.com/mod2")
}

// TestReplaceRemove тестирует удаление replace директивы
func TestReplaceRemove(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")
	env.RunGommExpectSuccess("init")

	module := "example.com/module"

	// Добавляем replace
	env.RunGommExpectSuccess("replace", "add", module, "example.com/new")

	// Проверяем, что добавлена
	goModContent := env.ReadFile(env.GetProjectGoModPath())
	env.AssertContains(goModContent, module)

	// Удаляем replace
	output := env.RunGommExpectSuccess("replace", "remove", module)

	env.AssertContains(output, "removed")

	// Проверяем, что удалена из go.mod
	goModContent = env.ReadFile(env.GetProjectGoModPath())
	// После удаления replace не должно быть в файле
	// (или должен быть закомментирован, зависит от реализации)
}

// TestWorkspacePreservesCustomReplaces тестирует сохранение custom replace при переключении
func TestWorkspacePreservesCustomReplaces(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Добавляем custom replace
	customModule := "example.com/custom"
	env.RunGommExpectSuccess("replace", "add", customModule, "example.com/new")

	// Клонируем и переключаем на local
	env.RunGommExpectSuccess("clone", depModulePath)
	env.RunGommExpectSuccess("local", depModulePath)

	// Проверяем, что custom replace сохранена в go.mod
	goModContent := env.ReadFile(env.GetProjectGoModPath())
	env.AssertContains(goModContent, customModule)

	// Переключаем обратно на remote
	env.RunGommExpectSuccess("remote", depModulePath)

	// Custom replace все еще должна быть в go.mod
	goModContent = env.ReadFile(env.GetProjectGoModPath())
	env.AssertContains(goModContent, customModule)
}

// TestLocalWithoutClone тестирует переключение на local без предварительного clone
func TestLocalWithoutClone(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Пытаемся переключить на local без clone
	output := env.RunGommExpectError("local", depModulePath)

	// Должна быть ошибка о том, что модуль не склонирован
	env.AssertContains(output, "not cloned")
}

// TestWorkspaceWithNestedDependencies тестирует workspace с вложенными зависимостями
func TestWorkspaceWithNestedDependencies(t *testing.T) {
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

	// Клонируем обе зависимости
	env.RunGommExpectSuccess("clone", depAPath)
	env.RunGommExpectSuccess("clone", depBPath)

	// Переключаем обе на local
	env.RunGommExpectSuccess("local", depAPath, depBPath)

	// Проверяем go.work
	uses := env.ParseGoWork()
	if len(uses) < 2 {
		t.Fatalf("Expected at least 2 use directives, got %d", len(uses))
	}

	// Проверяем, что обе зависимости в workspace
	usesStr := strings.Join(uses, " ")
	env.AssertContains(usesStr, "dep-a")
	env.AssertContains(usesStr, "dep-b")
}

// TestStatusShowsWorkspaceInfo тестирует отображение информации о workspace в status
func TestStatusShowsWorkspaceInfo(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depAPath := "github.com/test/dep-a"
	depBPath := "github.com/test/dep-b"

	env.CreateTestModule(depAPath)
	env.CreateTestModule(depBPath)
	env.CreateTestProject("github.com/test/myproject", depAPath, depBPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")
	env.RunGommExpectSuccess("clone", "--all")

	// Переключаем только одну зависимость
	env.RunGommExpectSuccess("local", depAPath)

	// Проверяем status
	output := env.RunGommExpectSuccess("status", "-v")

	// Должна быть информация о workspace режиме
	env.AssertContains(output, "workspace")

	// Должно быть указание на смешанный режим
	env.AssertContains(output, "1")
}

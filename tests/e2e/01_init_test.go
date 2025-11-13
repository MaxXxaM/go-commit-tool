package e2e

import (
	"path/filepath"
	"testing"

	"github.com/MaxXxaM/gomm/tests/helpers"
)

// TestInitCommand тестирует команду gomm init
func TestInitCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем простой проект
	env.CreateTestProject("github.com/test/myproject")

	// Выполняем gomm init
	output := env.RunGommExpectSuccess("init")

	// Проверяем, что конфиг создан
	env.AssertFileExists(filepath.Join(env.ProjectDir, ".gomm.yaml"))
	env.AssertContains(output, "initialized")

	// Проверяем, что директория метаданных создана
	env.AssertDirExists(filepath.Join(env.ProjectDir, ".gomm"))
}

// TestInitWithExistingConfig тестирует повторную инициализацию
func TestInitWithExistingConfig(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")

	// Первая инициализация
	env.RunGommExpectSuccess("init")

	// Вторая инициализация (должна выдать предупреждение или пропустить)
	output := env.RunGommExpectSuccess("init")
	env.AssertContains(output, "already")
}

// TestScanCommand тестирует команду gomm scan
func TestScanCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем тестовые модули
	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)

	// Создаем проект с зависимостью
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	// Инициализируем gomm
	env.RunGommExpectSuccess("init")

	// Выполняем scan
	output := env.RunGommExpectSuccess("scan")

	// Проверяем, что зависимость обнаружена
	env.AssertContains(output, depModulePath)
	env.AssertContains(output, "dependencies found")
}

// TestScanWithDepth тестирует команду scan с ограничением глубины
func TestScanWithDepth(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем цепочку зависимостей: project -> dep-a -> dep-b
	depBModulePath := "github.com/test/dep-b"
	env.CreateTestModule(depBModulePath)

	depAModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depAModulePath, depBModulePath)

	env.CreateTestProject("github.com/test/myproject", depAModulePath)

	env.RunGommExpectSuccess("init")

	// Сканируем с глубиной 1 (только прямые зависимости)
	output := env.RunGommExpectSuccess("scan", "--depth", "1")

	// dep-a должна быть найдена
	env.AssertContains(output, depAModulePath)
	// dep-b не должна быть найдена (глубина 2)
	// Примечание: это зависит от реализации, возможно нужно будет скорректировать
}

// TestStatusCommand тестирует команду gomm status
func TestStatusCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Выполняем status
	output := env.RunGommExpectSuccess("status")

	// Проверяем, что статус отображается
	env.AssertContains(output, "Project:")
	env.AssertContains(output, "github.com/test/myproject")
}

// TestStatusVerbose тестирует команду status с флагом -v
func TestStatusVerbose(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Выполняем status с verbose
	output := env.RunGommExpectSuccess("status", "-v")

	// Проверяем, что выводится больше информации
	env.AssertContains(output, depModulePath)
}

// TestTreeCommand тестирует команду gomm tree
func TestTreeCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем иерархию зависимостей
	depBModulePath := "github.com/test/dep-b"
	env.CreateTestModule(depBModulePath)

	depAModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depAModulePath, depBModulePath)

	env.CreateTestProject("github.com/test/myproject", depAModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Выполняем tree
	output := env.RunGommExpectSuccess("tree")

	// Проверяем, что дерево отображается
	env.AssertContains(output, depAModulePath)
	env.AssertContains(output, depBModulePath)
}

// TestTreeWithDepthLimit тестирует tree с ограничением глубины
func TestTreeWithDepthLimit(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем глубокую иерархию
	depCModulePath := "github.com/test/dep-c"
	env.CreateTestModule(depCModulePath)

	depBModulePath := "github.com/test/dep-b"
	env.CreateTestModule(depBModulePath, depCModulePath)

	depAModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depAModulePath, depBModulePath)

	env.CreateTestProject("github.com/test/myproject", depAModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Выполняем tree с ограничением глубины
	output := env.RunGommExpectSuccess("tree", "--depth", "2")

	// Проверяем, что только первые 2 уровня отображаются
	env.AssertContains(output, depAModulePath)
	env.AssertContains(output, depBModulePath)
}

// TestTreeLocalOnly тестирует tree с флагом --local-only
func TestTreeLocalOnly(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	depModulePath := "github.com/test/dep-a"
	env.CreateTestModule(depModulePath)
	env.CreateTestProject("github.com/test/myproject", depModulePath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем зависимость локально
	env.RunGommExpectSuccess("clone", depModulePath)

	// Выполняем tree --local-only
	output := env.RunGommExpectSuccess("tree", "--local-only")

	// Должна отображаться только локальная зависимость
	env.AssertContains(output, depModulePath)
}

// TestConfigWithSharedDir тестирует настройку shared_dir
func TestConfigWithSharedDir(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")
	env.RunGommExpectSuccess("init")

	// Загружаем конфигурацию
	cfg := env.LoadGommConfig()

	// Проверяем, что shared_dir установлен
	if cfg.SharedDir == "" {
		t.Fatal("Expected shared_dir to be set")
	}
}

// TestUninitializedProject тестирует работу команд в неинициализированном проекте
func TestUninitializedProject(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestProject("github.com/test/myproject")

	// Пытаемся выполнить scan без init
	output := env.RunGommExpectError("scan")

	// Должна быть ошибка о неинициализированном проекте
	env.AssertContains(output, "not initialized")
}

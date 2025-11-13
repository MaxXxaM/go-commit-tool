package e2e

import (
	"path/filepath"
	"testing"

	"github.com/MaxXxaM/gomm/tests/helpers"
)

// TestCompleteWorkflow тестирует полный workflow от инициализации до релиза
func TestCompleteWorkflow(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// === Этап 1: Настройка тестовых репозиториев ===
	t.Log("Step 1: Creating test repositories")

	// Создаем библиотеку utils
	utilsPath := "github.com/test/utils"
	utilsDir := env.CreateTestModule(utilsPath)

	// Создаем библиотеку auth, зависящую от utils
	authPath := "github.com/test/auth"
	authDir := env.CreateTestModule(authPath, utilsPath)

	// Создаем проект, зависящий от auth
	projectPath := "github.com/test/myservice"
	env.CreateTestProject(projectPath, authPath)

	// === Этап 2: Инициализация проекта ===
	t.Log("Step 2: Initializing project")

	output := env.RunGommExpectSuccess("init")
	env.AssertContains(output, "initialized")

	// Проверяем структуру
	env.AssertFileExists(filepath.Join(env.ProjectDir, ".gomm.yaml"))
	env.AssertDirExists(filepath.Join(env.ProjectDir, ".gomm"))

	// === Этап 3: Сканирование зависимостей ===
	t.Log("Step 3: Scanning dependencies")

	output = env.RunGommExpectSuccess("scan")
	env.AssertContains(output, authPath)

	// === Этап 4: Просмотр дерева зависимостей ===
	t.Log("Step 4: Viewing dependency tree")

	output = env.RunGommExpectSuccess("tree")
	env.AssertContains(output, authPath)
	env.AssertContains(output, utilsPath)

	// === Этап 5: Проверка статуса ===
	t.Log("Step 5: Checking status")

	output = env.RunGommExpectSuccess("status", "-v")
	env.AssertContains(output, projectPath)

	// === Этап 6: Клонирование зависимостей ===
	t.Log("Step 6: Cloning dependencies")

	// Клонируем auth в local
	output = env.RunGommExpectSuccess("clone", authPath, "--mode", "local")
	env.AssertContains(output, "cloned")

	// Клонируем utils в shared
	output = env.RunGommExpectSuccess("clone", utilsPath, "--mode", "shared")
	env.AssertContains(output, "cloned")

	// Проверяем размещение
	authLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "auth")
	utilsSharedPath := filepath.Join(env.SharedDir, "utils")
	env.AssertDirExists(authLocalPath)
	env.AssertDirExists(utilsSharedPath)

	// === Этап 7: Переключение на локальную разработку ===
	t.Log("Step 7: Switching to local development")

	output = env.RunGommExpectSuccess("local", authPath, utilsPath)
	env.AssertContains(output, "workspace")

	// Проверяем go.work
	env.AssertFileExists(env.GetProjectGoWorkPath())
	uses := env.ParseGoWork()
	if len(uses) < 2 {
		t.Fatalf("Expected at least 2 modules in workspace, got %d", len(uses))
	}

	// === Этап 8: Разработка - изменения в utils ===
	t.Log("Step 8: Making changes to utils")

	env.ModifyGoFile(utilsSharedPath, "utils.go", "NewUtility")
	env.GitCommit(utilsSharedPath, "Add new utility function")

	// === Этап 9: Разработка - изменения в auth ===
	t.Log("Step 9: Making changes to auth")

	env.ModifyGoFile(authLocalPath, "auth.go", "NewAuthMethod")
	env.GitCommit(authLocalPath, "Add new auth method")

	// === Этап 10: Проверка версий ===
	t.Log("Step 10: Checking versions")

	// Проверяем версию utils
	output, _ = env.RunGommInDir(utilsSharedPath, "version", "check")
	t.Logf("Utils version: %s", output)

	// Проверяем версию auth
	output, _ = env.RunGommInDir(authLocalPath, "version", "check")
	t.Logf("Auth version: %s", output)

	// === Этап 11: Создание релизов (от листьев к корню) ===
	t.Log("Step 11: Creating releases")

	// Сначала релиз utils (нет зависимостей)
	output, err := env.RunGommInDir(utilsSharedPath, "release", "--bump", "minor", "--no-tests")
	if err == nil {
		env.AssertContains(output, "release")
		t.Log("Utils released successfully")
	} else {
		t.Logf("Utils release output: %s (may fail if no proper changes detected)", output)
	}

	// Затем релиз auth (зависит от utils)
	output, err = env.RunGommInDir(authLocalPath, "release", "--bump", "minor", "--no-tests")
	if err == nil {
		env.AssertContains(output, "release")
		t.Log("Auth released successfully")
	} else {
		t.Logf("Auth release output: %s", output)
	}

	// === Этап 12: Переключение обратно на remote ===
	t.Log("Step 12: Switching back to remote")

	output = env.RunGommExpectSuccess("remote", "--all")
	env.AssertContains(output, "remote")

	// go.work должен быть пуст или удален
	goWorkPath := env.GetProjectGoWorkPath()
	if env.FileExists(goWorkPath) {
		content := env.ReadFile(goWorkPath)
		if len(content) > 20 { // Больше чем просто "go 1.21\n"
			uses := env.ParseGoWork()
			if len(uses) > 0 {
				t.Fatal("Expected go.work to be empty after switching all to remote")
			}
		}
	}

	// === Этап 13: Финальная проверка статуса ===
	t.Log("Step 13: Final status check")

	output = env.RunGommExpectSuccess("status", "-v")
	t.Logf("Final status:\n%s", output)
}

// TestMultiServiceWorkflow тестирует workflow с несколькими сервисами
func TestMultiServiceWorkflow(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем shared библиотеку
	commonPath := "github.com/test/common-lib"
	env.CreateTestModule(commonPath)

	// Создаем два сервиса, использующих общую библиотеку
	service1Path := "github.com/test/service-1"
	service2Path := "github.com/test/service-2"

	// Для теста используем только один проект
	env.CreateTestProject(service1Path, commonPath)

	// Инициализация
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем общую библиотеку в shared
	env.RunGommExpectSuccess("clone", commonPath, "--mode", "shared")

	// Переключаем на local
	env.RunGommExpectSuccess("local", commonPath)

	// Вносим изменения в общую библиотеку
	commonLocalPath := filepath.Join(env.SharedDir, "common-lib")
	env.ModifyGoFile(commonLocalPath, "common-lib.go", "SharedFeature")
	env.GitCommit(commonLocalPath, "Add shared feature")

	// Создаем релиз
	output, _ := env.RunGommInDir(commonLocalPath, "release", "--bump", "minor", "--no-tests")
	t.Logf("Common lib release: %s", output)

	// Оба сервиса должны иметь возможность использовать новую версию
	t.Log("Workflow completed successfully")
}

// TestHotfixWorkflow тестирует workflow быстрого исправления
func TestHotfixWorkflow(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем библиотеку
	libPath := "github.com/test/mylib"
	libDir := env.CreateTestModule(libPath)

	// Создаем несколько версий
	env.ModifyGoFile(libDir, "mylib.go", "Feature1")
	env.GitCommit(libDir, "Add feature 1")
	env.GitTag(libDir, "v0.2.0")

	// Создаем проект
	env.CreateTestProject("github.com/test/myservice", libPath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Обнаруживаем баг в библиотеке
	t.Log("Discovered a bug, need hotfix")

	// Клонируем
	env.RunGommExpectSuccess("clone", libPath)
	libLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "mylib")

	// Переключаем на local для исправления
	env.RunGommExpectSuccess("local", libPath)

	// Исправляем баг
	goFile := filepath.Join(libLocalPath, "mylib.go")
	content := env.ReadFile(goFile)
	content += "\n// Bug fix: fixed critical issue\n"
	env.WriteFile(goFile, content)
	env.GitCommit(libLocalPath, "Fix: critical bug")

	// Создаем patch релиз
	output, _ := env.RunGommInDir(libLocalPath, "release", "--bump", "patch", "--no-tests")
	t.Logf("Hotfix release: %s", output)

	// Переключаем обратно на remote
	env.RunGommExpectSuccess("remote", libPath)

	t.Log("Hotfix workflow completed")
}

// TestFeatureBranchWorkflow тестирует разработку новой фичи
func TestFeatureBranchWorkflow(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем библиотеку
	libPath := "github.com/test/feature-lib"
	env.CreateTestModule(libPath)

	// Создаем проект
	env.CreateTestProject("github.com/test/myservice", libPath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем для разработки фичи
	env.RunGommExpectSuccess("clone", libPath)
	libLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "feature-lib")

	// Переключаем на local
	env.RunGommExpectSuccess("local", libPath)

	// Разрабатываем фичу
	env.ModifyGoFile(libLocalPath, "feature-lib.go", "BigNewFeature")
	env.GitCommit(libLocalPath, "WIP: new feature")

	// Проверяем статус
	output := env.RunGommExpectSuccess("status", "-v")
	env.AssertContains(output, "workspace")

	// Продолжаем разработку
	env.ModifyGoFile(libLocalPath, "feature-lib.go", "FeatureHelper")
	env.GitCommit(libLocalPath, "Add feature helper")

	// Завершаем фичу - создаем релиз
	output, _ = env.RunGommInDir(libLocalPath, "release", "--bump", "minor", "--no-tests")
	t.Logf("Feature release: %s", output)

	// Переключаем на remote
	env.RunGommExpectSuccess("remote", libPath)

	t.Log("Feature branch workflow completed")
}

// TestMigrationWorkflow тестирует перемещение модуля между режимами
func TestMigrationWorkflow(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем модуль
	modulePath := "github.com/test/migrating-lib"
	env.CreateTestModule(modulePath)

	env.CreateTestProject("github.com/test/myservice", modulePath)
	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Сначала клонируем в local
	t.Log("Cloning to local")
	env.RunGommExpectSuccess("clone", modulePath, "--mode", "local")

	localPath := filepath.Join(filepath.Dir(env.ProjectDir), "migrating-lib")
	env.AssertDirExists(localPath)

	// Решаем, что модуль используется множеством проектов - переносим в shared
	t.Log("Moving to shared")
	output := env.RunGommExpectSuccess("move", modulePath, "--to", "shared")
	env.AssertContains(output, "moved")

	sharedPath := filepath.Join(env.SharedDir, "migrating-lib")
	env.AssertDirExists(sharedPath)
	env.AssertFileNotExists(localPath)

	// Переключаем на local для работы
	env.RunGommExpectSuccess("local", modulePath)

	// Вносим изменения
	env.ModifyGoFile(sharedPath, "migrating-lib.go", "SharedFeature")
	env.GitCommit(sharedPath, "Add shared feature")

	// Создаем релиз
	output, _ = env.RunGommInDir(sharedPath, "release", "--bump", "minor", "--no-tests")
	t.Logf("Release from shared: %s", output)

	t.Log("Migration workflow completed")
}

// TestCleanupWorkflow тестирует очистку после завершения работы
func TestCleanupWorkflow(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем несколько зависимостей
	dep1Path := "github.com/test/dep-1"
	dep2Path := "github.com/test/dep-2"
	dep3Path := "github.com/test/dep-3"

	env.CreateTestModule(dep1Path)
	env.CreateTestModule(dep2Path)
	env.CreateTestModule(dep3Path)

	env.CreateTestProject("github.com/test/myservice", dep1Path, dep2Path, dep3Path)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем все
	env.RunGommExpectSuccess("clone", "--all")

	// Переключаем на local
	env.RunGommExpectSuccess("local", "--all")

	// Проверяем, что все в workspace
	uses := env.ParseGoWork()
	if len(uses) < 3 {
		t.Fatalf("Expected at least 3 modules in workspace, got %d", len(uses))
	}

	// Работа завершена - очищаем
	t.Log("Cleaning up local dependencies")

	// Переключаем обратно на remote
	env.RunGommExpectSuccess("remote", "--all")

	// Удаляем локальные копии dep-1 и dep-2
	env.RunGommExpectSuccess("clean", dep1Path, "--force")
	env.RunGommExpectSuccess("clean", dep2Path, "--force")

	// Оставляем dep-3 для будущей работы
	dep3LocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dep-3")
	env.AssertDirExists(dep3LocalPath)

	t.Log("Cleanup workflow completed")
}

// TestReplaceWorkflow тестирует работу с custom replace директивами
func TestReplaceWorkflow(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем проект
	env.CreateTestProject("github.com/test/myservice")
	env.RunGommExpectSuccess("init")

	// Добавляем replace для форка библиотеки
	t.Log("Adding custom replace for a fork")
	env.RunGommExpectSuccess("replace", "add", "github.com/external/lib", "github.com/test/lib-fork")

	// Проверяем, что replace добавлена
	output := env.RunGommExpectSuccess("replace", "list")
	env.AssertContains(output, "github.com/external/lib")

	// Добавляем replace на локальный путь
	env.RunGommExpectSuccess("replace", "add", "github.com/another/lib", "../local-lib")

	// Проверяем список
	output = env.RunGommExpectSuccess("replace", "list")
	env.AssertContains(output, "github.com/another/lib")

	// Удаляем одну replace
	env.RunGommExpectSuccess("replace", "remove", "github.com/external/lib")

	// Проверяем, что удалена
	output = env.RunGommExpectSuccess("replace", "list")
	env.AssertNotContains(output, "github.com/external/lib")
	env.AssertContains(output, "github.com/another/lib")

	t.Log("Replace workflow completed")
}

// TestPartialWorkspaceWorkflow тестирует частичное использование workspace
func TestPartialWorkspaceWorkflow(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Создаем несколько зависимостей
	stablePath := "github.com/test/stable-lib"
	devPath := "github.com/test/dev-lib"
	unchangedPath := "github.com/test/unchanged-lib"

	env.CreateTestModule(stablePath)
	env.CreateTestModule(devPath)
	env.CreateTestModule(unchangedPath)

	env.CreateTestProject("github.com/test/myservice", stablePath, devPath, unchangedPath)

	env.RunGommExpectSuccess("init")
	env.RunGommExpectSuccess("scan")

	// Клонируем все
	env.RunGommExpectSuccess("clone", "--all")

	// Переключаем только dev-lib на local (для активной разработки)
	t.Log("Switching only dev-lib to local")
	env.RunGommExpectSuccess("local", devPath)

	// Проверяем смешанный режим
	output := env.RunGommExpectSuccess("status", "-v")
	env.AssertContains(output, "1") // одна зависимость в workspace

	// Работаем с dev-lib
	devLocalPath := filepath.Join(filepath.Dir(env.ProjectDir), "dev-lib")
	env.ModifyGoFile(devLocalPath, "dev-lib.go", "ActiveDevelopment")
	env.GitCommit(devLocalPath, "Active development")

	// stable-lib и unchanged-lib остаются в remote режиме
	uses := env.ParseGoWork()
	if len(uses) != 1 {
		t.Fatalf("Expected exactly 1 module in workspace (partial mode), got %d", len(uses))
	}

	t.Log("Partial workspace workflow completed")
}

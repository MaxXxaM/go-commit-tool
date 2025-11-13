package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MaxXxaM/gomm/internal/config"
	"gopkg.in/yaml.v3"
)

// TestEnv представляет тестовое окружение
type TestEnv struct {
	T               *testing.T
	TempDir         string
	ProjectDir      string
	SharedDir       string
	GommBinary      string
	OriginalDir     string
	TestReposDir    string
	CleanupFuncs    []func()
}

// NewTestEnv создает новое тестовое окружение
func NewTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	// Сохраняем текущую директорию
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Создаем временную директорию
	tempDir := t.TempDir()

	// Создаем структуру директорий
	projectDir := filepath.Join(tempDir, "project")
	sharedDir := filepath.Join(tempDir, "shared")
	testReposDir := filepath.Join(tempDir, "test-repos")

	for _, dir := range []string{projectDir, sharedDir, testReposDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Путь к бинарнику gomm
	gommBinary := os.Getenv("GOMM_BINARY")
	if gommBinary == "" {
		// Попытка найти в PATH
		gommBinary, err = exec.LookPath("gomm")
		if err != nil {
			// Пытаемся собрать из исходников
			rootDir := filepath.Join(originalDir, "..", "..")
			gommBinary = filepath.Join(rootDir, "gomm")

			cmd := exec.Command("go", "build", "-o", gommBinary, "./cmd/gomm")
			cmd.Dir = rootDir
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("Failed to build gomm: %v\nOutput: %s", err, output)
			}
		}
	}

	env := &TestEnv{
		T:            t,
		TempDir:      tempDir,
		ProjectDir:   projectDir,
		SharedDir:    sharedDir,
		GommBinary:   gommBinary,
		OriginalDir:  originalDir,
		TestReposDir: testReposDir,
		CleanupFuncs: []func(){},
	}

	return env
}

// Cleanup выполняет очистку тестового окружения
func (env *TestEnv) Cleanup() {
	// Возвращаемся в оригинальную директорию
	if err := os.Chdir(env.OriginalDir); err != nil {
		env.T.Logf("Failed to change back to original directory: %v", err)
	}

	// Выполняем дополнительные cleanup функции
	for _, cleanup := range env.CleanupFuncs {
		cleanup()
	}
}

// RunGomm выполняет команду gomm в проектной директории
func (env *TestEnv) RunGomm(args ...string) (string, error) {
	return env.RunGommInDir(env.ProjectDir, args...)
}

// RunGommInDir выполняет команду gomm в указанной директории
func (env *TestEnv) RunGommInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command(env.GommBinary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", env.TempDir),
	)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunGommExpectSuccess выполняет команду gomm и ожидает успех
func (env *TestEnv) RunGommExpectSuccess(args ...string) string {
	env.T.Helper()
	output, err := env.RunGomm(args...)
	if err != nil {
		env.T.Fatalf("gomm %s failed: %v\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return output
}

// RunGommExpectError выполняет команду gomm и ожидает ошибку
func (env *TestEnv) RunGommExpectError(args ...string) string {
	env.T.Helper()
	output, err := env.RunGomm(args...)
	if err == nil {
		env.T.Fatalf("gomm %s should have failed but succeeded\nOutput: %s", strings.Join(args, " "), output)
	}
	return output
}

// CreateTestModule создает тестовый Go модуль
func (env *TestEnv) CreateTestModule(modulePath string, deps ...string) string {
	env.T.Helper()

	// Создаем директорию для модуля
	moduleName := filepath.Base(modulePath)
	moduleDir := filepath.Join(env.TestReposDir, moduleName)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		env.T.Fatalf("Failed to create module directory: %v", err)
	}

	// Инициализируем git репозиторий
	env.RunCommand(moduleDir, "git", "init")
	env.RunCommand(moduleDir, "git", "config", "user.name", "Test User")
	env.RunCommand(moduleDir, "git", "config", "user.email", "test@example.com")

	// Создаем go.mod
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n", modulePath)
	if len(deps) > 0 {
		goModContent += "\nrequire (\n"
		for _, dep := range deps {
			goModContent += fmt.Sprintf("\t%s v0.1.0\n", dep)
		}
		goModContent += ")\n"
	}

	if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		env.T.Fatalf("Failed to write go.mod: %v", err)
	}

	// Создаем простой Go файл
	mainContent := fmt.Sprintf(`package %s

// Version returns the version of the module
func Version() string {
	return "0.1.0"
}

// Hello returns a greeting message
func Hello() string {
	return "Hello from %s"
}
`, moduleName, moduleName)

	if err := os.WriteFile(filepath.Join(moduleDir, moduleName+".go"), []byte(mainContent), 0644); err != nil {
		env.T.Fatalf("Failed to write Go file: %v", err)
	}

	// Создаем начальный коммит
	env.RunCommand(moduleDir, "git", "add", ".")
	env.RunCommand(moduleDir, "git", "commit", "-m", "Initial commit")
	env.RunCommand(moduleDir, "git", "tag", "v0.1.0")

	return moduleDir
}

// CreateTestProject создает тестовый проект с зависимостями
func (env *TestEnv) CreateTestProject(modulePath string, deps ...string) {
	env.T.Helper()

	// Инициализируем git репозиторий
	env.RunCommand(env.ProjectDir, "git", "init")
	env.RunCommand(env.ProjectDir, "git", "config", "user.name", "Test User")
	env.RunCommand(env.ProjectDir, "git", "config", "user.email", "test@example.com")

	// Создаем go.mod
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n", modulePath)
	if len(deps) > 0 {
		goModContent += "\nrequire (\n"
		for _, dep := range deps {
			goModContent += fmt.Sprintf("\t%s v0.1.0\n", dep)
		}
		goModContent += ")\n"
	}

	if err := os.WriteFile(filepath.Join(env.ProjectDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		env.T.Fatalf("Failed to write go.mod: %v", err)
	}

	// Создаем main.go
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Test project")
}
`
	if err := os.WriteFile(filepath.Join(env.ProjectDir, "main.go"), []byte(mainContent), 0644); err != nil {
		env.T.Fatalf("Failed to write main.go: %v", err)
	}

	// Создаем начальный коммит
	env.RunCommand(env.ProjectDir, "git", "add", ".")
	env.RunCommand(env.ProjectDir, "git", "commit", "-m", "Initial commit")
}

// RunCommand выполняет команду в указанной директории
func (env *TestEnv) RunCommand(dir string, name string, args ...string) string {
	env.T.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		env.T.Fatalf("Command %s %s failed in %s: %v\nOutput: %s", name, strings.Join(args, " "), dir, err, output)
	}
	return string(output)
}

// FileExists проверяет существование файла
func (env *TestEnv) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists проверяет существование директории
func (env *TestEnv) DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ReadFile читает содержимое файла
func (env *TestEnv) ReadFile(path string) string {
	env.T.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		env.T.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}

// WriteFile записывает содержимое в файл
func (env *TestEnv) WriteFile(path string, content string) {
	env.T.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		env.T.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// AssertFileExists проверяет, что файл существует
func (env *TestEnv) AssertFileExists(path string) {
	env.T.Helper()
	if !env.FileExists(path) {
		env.T.Fatalf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists проверяет, что файл не существует
func (env *TestEnv) AssertFileNotExists(path string) {
	env.T.Helper()
	if env.FileExists(path) {
		env.T.Fatalf("Expected file to not exist: %s", path)
	}
}

// AssertDirExists проверяет, что директория существует
func (env *TestEnv) AssertDirExists(path string) {
	env.T.Helper()
	if !env.DirExists(path) {
		env.T.Fatalf("Expected directory to exist: %s", path)
	}
}

// AssertContains проверяет, что строка содержит подстроку
func (env *TestEnv) AssertContains(str, substr string) {
	env.T.Helper()
	if !strings.Contains(str, substr) {
		env.T.Fatalf("Expected string to contain %q, got: %s", substr, str)
	}
}

// AssertNotContains проверяет, что строка не содержит подстроку
func (env *TestEnv) AssertNotContains(str, substr string) {
	env.T.Helper()
	if strings.Contains(str, substr) {
		env.T.Fatalf("Expected string to not contain %q, got: %s", substr, str)
	}
}

// LoadGommConfig загружает конфигурацию gomm
func (env *TestEnv) LoadGommConfig() *config.Config {
	env.T.Helper()
	configPath := filepath.Join(env.ProjectDir, ".gomm.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		env.T.Fatalf("Failed to read config: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		env.T.Fatalf("Failed to parse config: %v", err)
	}
	return &cfg
}

// LoadMetadata загружает метаданные gomm
func (env *TestEnv) LoadMetadata() map[string]interface{} {
	env.T.Helper()
	metadataPath := filepath.Join(env.ProjectDir, ".gomm", "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		env.T.Fatalf("Failed to read metadata: %v", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		env.T.Fatalf("Failed to parse metadata: %v", err)
	}
	return metadata
}

// GitCommit создает git коммит в указанной директории
func (env *TestEnv) GitCommit(dir, message string) {
	env.T.Helper()
	env.RunCommand(dir, "git", "add", ".")
	env.RunCommand(dir, "git", "commit", "-m", message)
}

// GitTag создает git тег в указанной директории
func (env *TestEnv) GitTag(dir, tag string) {
	env.T.Helper()
	env.RunCommand(dir, "git", "tag", tag)
}

// ModifyGoFile изменяет Go файл (добавляет новую функцию)
func (env *TestEnv) ModifyGoFile(dir, filename, functionName string) {
	env.T.Helper()
	filePath := filepath.Join(dir, filename)

	// Добавляем новую функцию
	newFunction := fmt.Sprintf(`

// %s is a new function
func %s() string {
	return "%s"
}
`, functionName, functionName, functionName)

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		env.T.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(newFunction); err != nil {
		env.T.Fatalf("Failed to write to file: %v", err)
	}
}

// GetProjectGoModPath возвращает путь к go.mod проекта
func (env *TestEnv) GetProjectGoModPath() string {
	return filepath.Join(env.ProjectDir, "go.mod")
}

// GetProjectGoWorkPath возвращает путь к go.work проекта
func (env *TestEnv) GetProjectGoWorkPath() string {
	return filepath.Join(env.ProjectDir, "go.work")
}

// ParseGoWork парсит go.work файл и возвращает список use директив
func (env *TestEnv) ParseGoWork() []string {
	env.T.Helper()
	content := env.ReadFile(env.GetProjectGoWorkPath())

	var uses []string
	lines := strings.Split(content, "\n")
	inUse := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "use (") {
			inUse = true
			continue
		}
		if inUse && line == ")" {
			break
		}
		if inUse && line != "" && !strings.HasPrefix(line, "//") {
			uses = append(uses, strings.Trim(line, "\""))
		}
	}

	return uses
}

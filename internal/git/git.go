package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MaxXxaM/gomm/pkg/logger"
)

// Client представляет клиент для работы с Git
type Client struct {
	log *logger.Logger
}

// NewClient создает новый Git клиент
func NewClient(log *logger.Logger) *Client {
	return &Client{
		log: log,
	}
}

// Clone клонирует репозиторий в указанную директорию
func (c *Client) Clone(repoURL, targetDir string) error {
	c.log.Debug("Клонирование %s в %s", repoURL, targetDir)

	// Проверяем существует ли директория
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("директория %s уже существует", targetDir)
	}

	// Создаем родительскую директорию если нужно
	parentDir := filepath.Dir(targetDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", parentDir, err)
	}

	cmd := exec.Command("git", "clone", repoURL, targetDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка клонирования: %w\nВывод: %s", err, string(output))
	}

	c.log.Debug("Клонирование завершено: %s", string(output))
	return nil
}

// IsRepo проверяет, является ли директория git репозиторием
func (c *Client) IsRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// HasUncommittedChanges проверяет наличие незакоммиченных изменений
func (c *Client) HasUncommittedChanges(dir string) (bool, error) {
	cmd := exec.Command("git", "-C", dir, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("не удалось проверить статус: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// GetCurrentTag возвращает текущий тег (версию)
func (c *Client) GetCurrentTag(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "describe", "--tags", "--exact-match")
	output, err := cmd.Output()
	if err != nil {
		// Если нет точного тега, пытаемся получить последний тег
		cmd = exec.Command("git", "-C", dir, "describe", "--tags", "--abbrev=0")
		output, err = cmd.Output()
		if err != nil {
			return "", nil // Нет тегов
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// GetLatestTag возвращает последний тег в репозитории
func (c *Client) GetLatestTag(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", nil // Нет тегов
	}

	return strings.TrimSpace(string(output)), nil
}

// GetAllTags возвращает все теги в репозитории
func (c *Client) GetAllTags(dir string) ([]string, error) {
	cmd := exec.Command("git", "-C", dir, "tag", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить теги: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 1 && tags[0] == "" {
		return []string{}, nil
	}

	return tags, nil
}

// GetRemoteURL возвращает URL удаленного репозитория
func (c *Client) GetRemoteURL(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("не удалось получить URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch возвращает текущую ветку
func (c *Client) GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("не удалось получить ветку: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Fetch выполняет git fetch
func (c *Client) Fetch(dir string) error {
	cmd := exec.Command("git", "-C", dir, "fetch", "--all", "--tags")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("не удалось выполнить fetch: %w\nВывод: %s", err, string(output))
	}

	c.log.Debug("Fetch выполнен: %s", string(output))
	return nil
}

// Pull выполняет git pull
func (c *Client) Pull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("не удалось выполнить pull: %w\nВывод: %s", err, string(output))
	}

	c.log.Debug("Pull выполнен: %s", string(output))
	return nil
}

// Checkout выполняет checkout ветки или тега
func (c *Client) Checkout(dir, ref string) error {
	cmd := exec.Command("git", "-C", dir, "checkout", ref)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("не удалось выполнить checkout %s: %w\nВывод: %s", ref, err, string(output))
	}

	c.log.Debug("Checkout выполнен: %s", string(output))
	return nil
}

// ModulePathToRepoURL конвертирует module path в repository URL
func ModulePathToRepoURL(modulePath string) string {
	// Для простоты используем HTTPS
	// TODO: добавить поддержку SSH и других протоколов
	return "https://" + modulePath + ".git"
}

// IsGitInstalled проверяет, установлен ли git
func IsGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}

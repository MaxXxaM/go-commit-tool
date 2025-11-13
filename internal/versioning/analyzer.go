package versioning

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/MaxXxaM/gomm/internal/git"
	"github.com/MaxXxaM/gomm/pkg/logger"
	"github.com/MaxXxaM/gomm/pkg/types"
)

// Analyzer анализирует изменения и определяет версии
type Analyzer struct {
	gitClient *git.Client
	log       *logger.Logger
}

// NewAnalyzer создает новый анализатор версий
func NewAnalyzer(log *logger.Logger) *Analyzer {
	return &Analyzer{
		gitClient: git.NewClient(log),
		log:       log,
	}
}

// GetCurrentVersion возвращает текущую версию модуля
func (a *Analyzer) GetCurrentVersion(modulePath string) (types.Version, error) {
	if modulePath == "" {
		return "", fmt.Errorf("путь к модулю не указан")
	}

	tag, err := a.gitClient.GetLatestTag(modulePath)
	if err != nil || tag == "" {
		// Нет тегов, начинаем с v0.1.0
		return "v0.1.0", nil
	}

	return types.Version(tag), nil
}

// ParseVersion парсит версию в формате semantic versioning
func ParseVersion(version types.Version) (major, minor, patch int, err error) {
	versionStr := string(version)

	// Удаляем префикс v если есть
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Регулярное выражение для semantic versioning
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) != 4 {
		return 0, 0, 0, fmt.Errorf("невалидный формат версии: %s", version)
	}

	major, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("невалидный major: %w", err)
	}

	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("невалидный minor: %w", err)
	}

	patch, err = strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("невалидный patch: %w", err)
	}

	return major, minor, patch, nil
}

// BumpVersion увеличивает версию согласно типу
func BumpVersion(version types.Version, bump types.VersionBump) (types.Version, error) {
	major, minor, patch, err := ParseVersion(version)
	if err != nil {
		return "", err
	}

	switch bump {
	case types.VersionBumpMajor:
		major++
		minor = 0
		patch = 0
	case types.VersionBumpMinor:
		minor++
		patch = 0
	case types.VersionBumpPatch:
		patch++
	default:
		return "", fmt.Errorf("неизвестный тип bump: %s", bump)
	}

	return types.Version(fmt.Sprintf("v%d.%d.%d", major, minor, patch)), nil
}

// AnalyzeChanges анализирует изменения и определяет тип версии
func (a *Analyzer) AnalyzeChanges(modulePath string) (types.VersionBump, []string, error) {
	// Получаем текущую версию
	currentVersion, err := a.GetCurrentVersion(modulePath)
	if err != nil {
		return "", nil, err
	}

	// Получаем изменения с последнего тега
	changes, err := a.getChangesSinceTag(modulePath, string(currentVersion))
	if err != nil {
		return "", nil, fmt.Errorf("не удалось получить изменения: %w", err)
	}

	if len(changes) == 0 {
		return types.VersionBumpPatch, []string{"No changes detected"}, nil
	}

	// Анализируем изменения для определения типа версии
	bump := a.determineVersionBump(changes)

	return bump, changes, nil
}

// getChangesSinceTag получает список изменений с момента последнего тега
func (a *Analyzer) getChangesSinceTag(modulePath, tag string) ([]string, error) {
	// TODO: Реализовать более детальный анализ через git diff и AST
	// Пока возвращаем простую проверку наличия изменений

	hasChanges, err := a.gitClient.HasUncommittedChanges(modulePath)
	if err != nil {
		return nil, err
	}

	if hasChanges {
		return []string{"Uncommitted changes detected"}, nil
	}

	// Проверяем коммиты с последнего тега
	// Это упрощенная версия, в реальности нужно использовать git log
	return []string{"Changes since last tag"}, nil
}

// determineVersionBump определяет тип версии на основе изменений
func (a *Analyzer) determineVersionBump(changes []string) types.VersionBump {
	// Простая эвристика на основе сообщений коммитов
	// TODO: Реализовать более продвинутый анализ через AST

	hasBreaking := false
	hasFeature := false

	for _, change := range changes {
		changeLower := strings.ToLower(change)

		// Проверяем на breaking changes
		if strings.Contains(changeLower, "breaking") ||
			strings.Contains(changeLower, "!:") ||
			strings.Contains(changeLower, "removed") ||
			strings.Contains(changeLower, "renamed") {
			hasBreaking = true
			break
		}

		// Проверяем на новые фичи
		if strings.Contains(changeLower, "feat") ||
			strings.Contains(changeLower, "feature") ||
			strings.Contains(changeLower, "add") ||
			strings.Contains(changeLower, "new") {
			hasFeature = true
		}
	}

	if hasBreaking {
		return types.VersionBumpMajor
	}

	if hasFeature {
		return types.VersionBumpMinor
	}

	return types.VersionBumpPatch
}

// SuggestNextVersion предлагает следующую версию на основе анализа
func (a *Analyzer) SuggestNextVersion(modulePath string) (types.Version, types.VersionBump, []string, error) {
	currentVersion, err := a.GetCurrentVersion(modulePath)
	if err != nil {
		return "", "", nil, err
	}

	bump, changes, err := a.AnalyzeChanges(modulePath)
	if err != nil {
		return "", "", nil, err
	}

	nextVersion, err := BumpVersion(currentVersion, bump)
	if err != nil {
		return "", "", nil, err
	}

	return nextVersion, bump, changes, nil
}

// ValidateVersion проверяет корректность версии
func ValidateVersion(version types.Version) error {
	_, _, _, err := ParseVersion(version)
	return err
}

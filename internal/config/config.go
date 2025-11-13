package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/viper"
)

const (
	// ConfigFileName - имя файла конфигурации
	ConfigFileName = ".gomm.yaml"
	// MetadataDir - директория для метаданных
	MetadataDir = ".gomm"
	// MetadataFileName - имя файла метаданных
	MetadataFileName = "metadata.json"
)

// LoadConfig загружает конфигурацию из файла
func LoadConfig(path string) (*types.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Значения по умолчанию
	v.SetDefault("shared_dir", "~/shared")
	v.SetDefault("default_placement", "auto")
	v.SetDefault("versioning.auto_detect", true)
	v.SetDefault("release.auto_push", true)
	v.SetDefault("release.create_changelog", false)
	v.SetDefault("release.run_tests", true)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("не удалось прочитать конфигурацию: %w", err)
	}

	var cfg types.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("не удалось распарсить конфигурацию: %w", err)
	}

	// Раскрываем ~ в путях
	if cfg.SharedDir != "" {
		cfg.SharedDir = expandHome(cfg.SharedDir)
	}

	return &cfg, nil
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(cfg *types.Config, path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Устанавливаем значения из структуры
	v.Set("shared_dir", cfg.SharedDir)
	v.Set("default_placement", cfg.DefaultPlacement)
	v.Set("placement_rules", cfg.PlacementRules)
	v.Set("auth", cfg.Auth)
	v.Set("versioning", cfg.Versioning)
	v.Set("release", cfg.Release)
	v.Set("exclude", cfg.Exclude)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("не удалось сохранить конфигурацию: %w", err)
	}

	return nil
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *types.Config {
	homeDir, _ := os.UserHomeDir()

	return &types.Config{
		SharedDir:        filepath.Join(homeDir, "shared"),
		DefaultPlacement: types.PlacementAuto,
		PlacementRules:   []types.PlacementRule{},
		Auth:             make(map[string]types.AuthConfig),
		Versioning: types.VersioningConfig{
			AutoDetect: true,
		},
		Release: types.ReleaseConfig{
			AutoPush:        true,
			CreateChangelog: false,
			RunTests:        true,
		},
		Exclude: []string{
			"golang.org/x/*",
			"github.com/golang/*",
			"std",
		},
	}
}

// expandHome заменяет ~ на домашнюю директорию
func expandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(homeDir, path[1:])
}

// GetMetadataPath возвращает путь к файлу метаданных
func GetMetadataPath() string {
	return filepath.Join(MetadataDir, MetadataFileName)
}

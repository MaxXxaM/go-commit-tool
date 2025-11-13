package cli

import (
	"fmt"

	"github.com/MaxXxaM/gomm/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	log     *logger.Logger
	verbose bool
)

// rootCmd представляет базовую команду
var rootCmd = &cobra.Command{
	Use:   "gomm",
	Short: "Go Multi-Module Development Utility",
	Long: `gomm - утилита для управления разработкой в мультимодульных Go проектах.

Позволяет:
- Управлять локальными и удаленными зависимостями
- Переключаться между режимами go.work и go.mod
- Автоматизировать версионирование и релизы
- Синхронизировать изменения между связанными модулями`,
	Version: "0.3.0",
}

// Execute выполняет root команду
func Execute(logger *logger.Logger) error {
	log = logger
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "подробный вывод")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose {
			log = logger.New(logger.LevelDebug, nil)
		}
	}
}

// getLogger возвращает текущий логгер
func getLogger() *logger.Logger {
	if log == nil {
		log = logger.Default
	}
	return log
}

// printError выводит ошибку и завершает выполнение
func printError(err error) {
	getLogger().ErrorMsg(fmt.Sprintf("%v", err))
}

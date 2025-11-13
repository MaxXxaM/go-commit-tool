package cli

import (
	"fmt"
	"os"

	"github.com/MaxXxaM/gomm/internal/config"
	"github.com/MaxXxaM/gomm/pkg/types"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Управление replace директивами",
	Long: `Управление кастомными replace директивами в go.mod.

Подкоманды:
  add     - добавить replace директиву
  remove  - удалить replace директиву
  list    - показать все replace директивы`,
}

var replaceAddCmd = &cobra.Command{
	Use:   "add <module> <replacement>",
	Short: "Добавить replace директиву",
	Long: `Добавляет replace директиву в go.mod.

Примеры:
  gomm replace add example.com/old example.com/new
  gomm replace add example.com/module ../local/path`,
	Args: cobra.ExactArgs(2),
	RunE: runReplaceAdd,
}

var replaceRemoveCmd = &cobra.Command{
	Use:   "remove <module>",
	Short: "Удалить replace директиву",
	Long: `Удаляет replace директиву из go.mod.

Примеры:
  gomm replace remove example.com/old`,
	Args: cobra.ExactArgs(1),
	RunE: runReplaceRemove,
}

var replaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "Показать все replace директивы",
	Long:  `Отображает список всех replace директив из go.mod.`,
	RunE:  runReplaceList,
}

func init() {
	rootCmd.AddCommand(replaceCmd)
	replaceCmd.AddCommand(replaceAddCmd)
	replaceCmd.AddCommand(replaceRemoveCmd)
	replaceCmd.AddCommand(replaceListCmd)
}

func runReplaceAdd(cmd *cobra.Command, args []string) error {
	log := getLogger()

	modulePath := args[0]
	replacement := args[1]

	// Читаем go.mod
	data, err := os.ReadFile("go.mod")
	if err != nil {
		printError(fmt.Errorf("не удалось прочитать go.mod: %w", err))
		return err
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		printError(fmt.Errorf("не удалось распарсить go.mod: %w", err))
		return err
	}

	// Добавляем replace
	if err := modFile.AddReplace(modulePath, "", replacement, ""); err != nil {
		printError(fmt.Errorf("не удалось добавить replace: %w", err))
		return err
	}

	// Сохраняем go.mod
	formatted := modfile.Format(modFile.Syntax)
	if err := os.WriteFile("go.mod", formatted, 0644); err != nil {
		printError(fmt.Errorf("не удалось сохранить go.mod: %w", err))
		return err
	}

	// Сохраняем в метаданные
	metadataMgr := config.NewMetadataManager()
	metadata, err := metadataMgr.LoadOrCreate()
	if err == nil {
		metadataMgr.AddCustomReplace(metadata, types.CustomReplace{
			Module:      types.ModulePath(modulePath),
			Replacement: replacement,
		})
		metadataMgr.Save(metadata)
	}

	log.Success("Replace директива добавлена")
	log.Info("  %s => %s", modulePath, replacement)

	return nil
}

func runReplaceRemove(cmd *cobra.Command, args []string) error {
	log := getLogger()

	modulePath := args[0]

	// Читаем go.mod
	data, err := os.ReadFile("go.mod")
	if err != nil {
		printError(fmt.Errorf("не удалось прочитать go.mod: %w", err))
		return err
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		printError(fmt.Errorf("не удалось распарсить go.mod: %w", err))
		return err
	}

	// Удаляем replace
	if err := modFile.DropReplace(modulePath, ""); err != nil {
		printError(fmt.Errorf("не удалось удалить replace: %w", err))
		return err
	}

	// Сохраняем go.mod
	formatted := modfile.Format(modFile.Syntax)
	if err := os.WriteFile("go.mod", formatted, 0644); err != nil {
		printError(fmt.Errorf("не удалось сохранить go.mod: %w", err))
		return err
	}

	// Удаляем из метаданных
	metadataMgr := config.NewMetadataManager()
	metadata, err := metadataMgr.LoadOrCreate()
	if err == nil {
		metadataMgr.RemoveCustomReplace(metadata, types.ModulePath(modulePath))
		metadataMgr.Save(metadata)
	}

	log.Success("Replace директива удалена")
	log.Info("  %s", modulePath)

	return nil
}

func runReplaceList(cmd *cobra.Command, args []string) error {
	log := getLogger()

	// Читаем go.mod
	data, err := os.ReadFile("go.mod")
	if err != nil {
		printError(fmt.Errorf("не удалось прочитать go.mod: %w", err))
		return err
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		printError(fmt.Errorf("не удалось распарсить go.mod: %w", err))
		return err
	}

	if len(modFile.Replace) == 0 {
		log.Info("Нет replace директив в go.mod")
		return nil
	}

	fmt.Println("Replace директивы:")
	fmt.Println()

	for _, r := range modFile.Replace {
		oldPath := r.Old.Path
		if r.Old.Version != "" {
			oldPath += " " + r.Old.Version
		}

		newPath := r.New.Path
		if r.New.Version != "" {
			newPath += " " + r.New.Version
		}

		fmt.Printf("  %s => %s\n", oldPath, newPath)
	}

	return nil
}

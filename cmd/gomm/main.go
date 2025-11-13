package main

import (
	"fmt"
	"os"

	"github.com/MaxXxaM/gomm/internal/cli"
	"github.com/MaxXxaM/gomm/pkg/logger"
)

func main() {
	log := logger.New(logger.LevelInfo, os.Stdout)

	if err := cli.Execute(log); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

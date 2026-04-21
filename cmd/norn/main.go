package main

import (
	"os"

	"github.com/mssantosdev/norn/internal/cli"
	"github.com/mssantosdev/norn/internal/ui/logger"
)

func main() {
	logger.Init(logger.Options{})
	if err := cli.Run(os.Args[1:]); err != nil {
		logger.Error("command failed", "error", err)
		os.Exit(1)
	}
}

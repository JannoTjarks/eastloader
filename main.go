package main

import (
	"jannotjarks/eastloader/cmd"

	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)
	cmd.Execute()
}

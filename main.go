package main

import (
	"log/slog"
	"os"
	"time"

	"rph/state"
	"rph/cmd"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable" // needed for windows :(
)

func main() {
	state.Setup()

	slog.SetDefault(slog.New(
		tint.NewHandler(colorable.NewColorable(os.Stdout), &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
		))

	cmd.Execute()
}

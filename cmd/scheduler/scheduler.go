package main

import (
	"os"

	"github.com/creekorful/trandoshan/internal/scheduler"
)

func main() {
	app := scheduler.GetApp()
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

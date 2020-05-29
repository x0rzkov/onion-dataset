package main

import (
	"os"

	"github.com/creekorful/trandoshan/internal/feeder"
)

func main() {
	app := feeder.GetApp()
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

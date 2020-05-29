package main

import (
	"os"

	"github.com/creekorful/trandoshan/internal/persister"
)

func main() {
	app := persister.GetApp()
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

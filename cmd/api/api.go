package main

import (
	"os"

	"github.com/creekorful/trandoshan/internal/api"
)

func main() {
	app := api.GetApp()
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

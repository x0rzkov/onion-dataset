package main

import (
	"os"

	"github.com/creekorful/trandoshan/internal/crawler"
)

func main() {
	app := crawler.GetApp()
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

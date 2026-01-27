// Package main is the entry point for clinvk CLI.
package main

import (
	"os"

	"github.com/signalridge/clinvoker/internal/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}

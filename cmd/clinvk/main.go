// Package main is the entry point for clinvk CLI.
package main

import (
	"errors"
	"os"

	"github.com/signalridge/clinvoker/internal/app"
)

var executeApp = app.Execute

func main() {
	if code := run(); code != 0 {
		os.Exit(code)
	}
}

func run() int {
	if err := executeApp(); err != nil {
		if errors.Is(err, app.ErrCommandTimeout) {
			return 6
		}
		return 1
	}
	return 0
}

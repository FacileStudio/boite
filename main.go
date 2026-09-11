package main

import (
	"os"

	"github.com/FacileStudio/boite/cmd"
)

// main forwards the exit code. fang already rendered the styled error, so a
// failed Execute is reported by its non-zero exit, never by a second print here.
func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

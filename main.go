package main

import (
	"os"

	"github.com/FacileStudio/boite/cmd"
	"github.com/FacileStudio/boite/cmd/egress"
)

// main forwards the exit code. fang already rendered the styled error, so a
// failed Execute is reported by its non-zero exit, never by a second print here.
func main() {
	if served, code := egress.MaybeServe(os.Args); served {
		os.Exit(code)
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/FacileStudio/boite/cmd"
)

var version = "dev"

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
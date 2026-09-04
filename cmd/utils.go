package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RunCommand runs a command and returns output
func RunCommand(args ...string) ([]byte, error) {
	// If multipass, try to find it in common locations
	if args[0] == "multipass" {
		if _, err := os.Stat("/snap/bin/multipass"); err == nil {
			fmt.Fprintf(os.Stderr, "Using multipass at /snap/bin/multipass\n")
			args[0] = "/snap/bin/multipass"
		}
	}
	cmd := exec.Command(args[0], args[1:]...)
	return cmd.CombinedOutput()
}

// checkCommand checks if a command is available and returns error if not
func checkCommand(name string) error {
	_, err := exec.LookPath(name)
	if err == nil {
		return nil
	}

	// Check common snap locations
	if name == "multipass" {
		snapPaths := []string{
			"/snap/bin/multipass",
			filepath.Join(os.Getenv("HOME"), "snap", "bin", "multipass"),
		}
		for _, p := range snapPaths {
			if _, err := os.Stat(p); err == nil {
				return nil
			}
		}
	}
	return err
}

// homeDir returns the home directory of the current user
func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// isRoot checks if the current user is root
func isRoot() bool {
	return os.Geteuid() == 0
}

// Spinner displays a simple spinner while a function runs
func Spinner(action string, f func() error) error {
	done := make(chan bool)
	go func() {
		runes := []rune{'-', '\\', '|', '/'}
		i := 0
		for {
			select {
			case <-done:
				fmt.Fprint(os.Stderr, "\r"+strings.Repeat(" ", len(action)+20)+"\r") // clear line
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s %c", action, runes[i%len(runes)])
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	err := f()
	done <- true
	return err
}
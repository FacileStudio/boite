package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func multipassPath() string {
	if p, err := exec.LookPath("multipass"); err == nil {
		return p
	}
	candidates := []string{
		"/snap/bin/multipass",
		filepath.Join(homeDir(), "snap", "bin", "multipass"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "multipass"
}

// RunCommand executes a command and returns its combined stdout and stderr.
func RunCommand(args ...string) ([]byte, error) {
	bin := args[0]
	if bin == "multipass" {
		bin = multipassPath()
	}
	cmd := exec.Command(bin, args[1:]...)
	return cmd.CombinedOutput()
}

func checkCommand(name string) error {
	if name == "multipass" {
		if p := multipassPath(); p != "multipass" {
			return nil
		}
	}
	_, err := exec.LookPath(name)
	return err
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return home
}

func isRoot() bool {
	return os.Geteuid() == 0
}

// Spinner displays an activity indicator while an action runs.
func Spinner(action string, f func() error) error {
	done := make(chan bool)
	go func() {
		runes := []rune{'-', '\\', '|', '/'}
		i := 0
		for {
			select {
			case <-done:
				fmt.Fprint(os.Stderr, "\r"+strings.Repeat(" ", len(action)+20)+"\r")
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

func runInteractive(args ...string) error {
	bin := args[0]
	if bin == "multipass" {
		bin = multipassPath()
	}
	cmd := exec.Command(bin, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ensureVMRunning(name string) error {
	if err := checkCommand("multipass"); err != nil {
		return fmt.Errorf("multipass not found in PATH")
	}

	output, err := RunCommand("multipass", "info", name)
	if err != nil {
		if strings.Contains(string(output), "not found") || strings.Contains(string(output), "no instance named") {
			return fmt.Errorf("VM '%s' not found. Create it first: boite create %s", name, name)
		}
		return fmt.Errorf("checking VM: %w", err)
	}

	if !strings.Contains(string(output), "Status: Running") {
		fmt.Printf("VM '%s' is not running, starting...\n", name)
		if _, err := RunCommand("multipass", "start", name); err != nil {
			return fmt.Errorf("starting VM: %w", err)
		}
	}
	return nil
}

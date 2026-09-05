package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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

func runCommand(args ...string) ([]byte, error) {
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

func spinner(action string, f func() error) error {
	done := make(chan bool)
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Fprint(os.Stderr, "\r"+strings.Repeat(" ", len(action)+20)+"\r")
				return
			default:
				frame := lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(frames[i%len(frames)])
				fmt.Fprintf(os.Stderr, "\r %s %s", frame, action)
				i++
				time.Sleep(80 * time.Millisecond)
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

	output, err := runCommand("multipass", "info", name)
	if err != nil {
		if strings.Contains(string(output), "not found") || strings.Contains(string(output), "no instance named") {
			return fmt.Errorf("VM '%s' not found. Create it first: boite create %s", name, name)
		}
		return fmt.Errorf("checking VM: %w", err)
	}

	if !strings.Contains(string(output), "Status: Running") {
		fmt.Printf("VM '%s' is not running, starting...\n", name)
		if _, err := runCommand("multipass", "start", name); err != nil {
			return fmt.Errorf("starting VM: %w", err)
		}
	}
	return nil
}

func syncWorkspaceZshrc(name string) {
	check := "[ -f /workspace/.zshrc_local ] && echo exists || echo missing"
	output, _ := runCommand("multipass", "exec", name, "--", "bash", "-c", check)
	if strings.TrimSpace(string(output)) == "exists" {
		cmdStr := "cp /workspace/.zshrc_local /home/boite/.zshrc && chown boite:boite /home/boite/.zshrc"
		runCommand("multipass", "exec", name, "--", "sudo", "bash", "-c", cmdStr)
	}
}

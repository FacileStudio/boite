package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return home
}

func RunCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	return cmd.CombinedOutput()
}

func checkCommand(name string) error {
	_, err := exec.LookPath(name)
	return err
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
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}


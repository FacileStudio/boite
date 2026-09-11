package cmd

import (
	"os"
	"os/exec"

	"charm.land/lipgloss/v2"
)

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return home
}

// RunCommand runs an external program and returns its combined output, along
// with any error the exec produced.
func RunCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	return cmd.CombinedOutput()
}

func checkCommand(name string) error {
	_, err := exec.LookPath(name)
	return err
}

// statusCellStyle returns the style for an instance status cell in the table
// render: green for running, dimmed for stopped, amber for anything else.
func statusCellStyle(status string) lipgloss.Style {
	st := lipgloss.NewStyle()
	switch status {
	case "running":
		st = st.Foreground(successColor).Bold(true)
	case "stopped":
		st = st.Foreground(subtleColor)
	default:
		st = st.Foreground(warnColor)
	}
	return st
}

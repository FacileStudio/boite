package cmd

import (
	"os"
	"os/exec"
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

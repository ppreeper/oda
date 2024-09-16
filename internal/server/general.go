package server

import (
	"fmt"
	"os"
	"os/exec"
)

func InstanceLogs() error {
	cmd := exec.Command("sudo",
		"journalctl",
		"-u",
		"odoo.service",
		"-f",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error getting logs %w", err)
	}
	return nil
}

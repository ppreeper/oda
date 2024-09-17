package server

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func odooService(action string) error {
	fmt.Println(action + " odoo service")
	cmd := exec.Command("sudo",
		"systemctl",
		action,
		"odoo.service",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("service odoo %s failed: %w", action, err)
	}
	return nil
}

func ServiceStart() error {
	return odooService("start")
}

func ServiceStop() error {
	return odooService("stop")
}

func ServiceRestart() error {
	ServiceStop()
	time.Sleep(2 * time.Second)
	ServiceStart()
	return nil
}

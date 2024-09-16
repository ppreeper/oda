package server

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func InstanceAppInstallUpgrade(install bool, modules ...string) error {
	iu := "-u"

	if install {
		iu = "-i"
	}

	mods := []string{}
	for _, mod := range modules {
		mm := strings.Split(mod, ",")
		mods = append(mods, mm...)
	}

	modList := strings.Join(mods, ",")

	cmd := exec.Command("odoo/odoo-bin",
		"--no-http", "--stop-after-init",
		"-c", "/opt/odoo/conf/odoo.conf",
		iu, modList,
	)
	cmd.Dir = "/opt/odoo"
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error installing/upgrading modules %w", err)
	}

	return nil
}

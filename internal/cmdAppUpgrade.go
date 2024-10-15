/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Short:   "Upgrade module(s)",
	Long:    `Upgrade module(s)`,
	GroupID: "app",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("no modules specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()
		instance, err := GetInstance(project)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get odoo instance %w", err)
			return
		}
		if instance == (Instance{}) || instance.State != "RUNNING" {
			fmt.Fprintln(os.Stderr, "no odoo instance running, please launch one first")
			return
		}

		uid, err := IncusGetUid(project, "odoo")
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get odoo uid %w", err)
		}

		incusCmd := exec.Command("incus", "exec", project, "--user", uid, "-t",
			"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
			"odoo/odoo-bin",
			"--no-http", "--stop-after-init",
			"-c", "/opt/odoo/conf/odoo.conf",
			"-u", moduleList(args...),
		)
		incusCmd.Stdin = os.Stdin
		incusCmd.Stdout = os.Stdout
		incusCmd.Stderr = os.Stderr
		if err := incusCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error installing/upgrading modules %w", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

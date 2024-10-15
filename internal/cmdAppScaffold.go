/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var scaffoldCmd = &cobra.Command{
	Use:     "scaffold",
	Short:   "Generates an Odoo module skeleton in addons",
	Long:    `Generates an Odoo module skeleton in addons`,
	GroupID: "app",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// watch for nested creation
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
			return
		}

		incusCmd := exec.Command("incus", "exec", project, "--user", uid, "-t",
			"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
			"odoo/odoo-bin",
			"scaffold", args[0], "/opt/odoo/addons",
		)
		incusCmd.Stdin = os.Stdin
		incusCmd.Stdout = os.Stdout
		incusCmd.Stderr = os.Stderr
		if err := incusCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error scaffolding module %w", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
}

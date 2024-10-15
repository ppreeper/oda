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

var username string

var execCmd = &cobra.Command{
	Use:     "exec",
	Short:   "Access the shell",
	Long:    `Access the shell`,
	GroupID: "instance",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()

		uid, err := IncusGetUid(project, username)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get odoo uid %w", err)
			return
		}

		incusCmd := exec.Command("incus", "exec", project, "--user", uid, "-t",
			"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
			"/bin/bash",
		)
		incusCmd.Stdin = os.Stdin
		incusCmd.Stdout = os.Stdout
		incusCmd.Stderr = os.Stderr
		if err := incusCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error executing", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().StringVarP(&username, "username", "u", "odoo", "username")
}

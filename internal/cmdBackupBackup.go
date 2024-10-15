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

var backupCmd = &cobra.Command{
	Use:     "backup",
	Short:   "Backup database filestore and addons",
	Long:    `Backup database filestore and addons`,
	GroupID: "backup",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()

		uid, err := IncusGetUid(project, "odoo")
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get odoo uid", err)
		}

		if err := exec.Command("incus", "exec", project, "--user", uid, "-t",
			"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
			"oda", "backup",
		).Run(); err != nil {
			fmt.Fprintln(os.Stderr, "backup failed", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

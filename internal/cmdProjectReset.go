/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var projectResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "reset project dir and db",
	Long:  `reset project dir and db`,
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		confim := AreYouSure("reset the project")
		if !confim {
			fmt.Fprintln(os.Stderr, "reset the project canceled")
			return
		}
		// stop
		_, project := GetProject()
		SetInstanceState(project, "stop")
		// instanceStatus := GetInstanceState(project)
		// fmt.Fprintln(os.Stderr, project, instanceStatus.Metadata.Status)

		// rm -rf data/*
		cwd, _ := GetProject()
		if err := RemoveContents(filepath.Join(cwd, "data")); err != nil {
			fmt.Fprintln(os.Stderr, "data files removal failed %w", err)
			return
		}

		// drop db
		dbhost := GetOdooConf(cwd, "db_host")
		dbname := GetOdooConf(cwd, "db_name")

		uid, err := IncusGetUid(dbhost, "postgres")
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get postgres user id: %w", err)
			return
		}
		if err := exec.Command("incus", "exec", dbhost, "--user", uid, "-t", "--",
			"dropdb", "--if-exists", "-U", "postgres", "-f", dbname,
		).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "could not drop postgresql database %s error: %v\n", dbname, err)
			return
		}
	},
}

func init() {
	projectCmd.AddCommand(projectResetCmd)
}

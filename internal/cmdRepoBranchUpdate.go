/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var repoBranchUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update Odoo branch repository",
	Long:  `update Odoo branch repository`,
	Run: func(cmd *cobra.Command, args []string) {
		versions := GetCurrentOdooRepos()
		versionOptions := []huh.Option[string]{}
		for _, version := range versions {
			versionOptions = append(versionOptions, huh.NewOption(version, version))
		}
		var version string
		var update bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Available Odoo Branches").
					Options(versionOptions...).
					Value(&version),

				huh.NewConfirm().
					Title("Update Branch?").
					Value(&update),
			),
		)
		if err := form.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "odoo version form error %v\n", err)
			return
		}
		if !update {
			return
		}

		repoDir := viper.GetString("dirs.repo")
		repos := []string{"odoo", "enterprise", "design-themes", "industry"}

		for _, repo := range repos {
			dest := filepath.Join(repoDir, version, repo)

			pull := exec.Command("git", "pull", "--rebase")
			pull.Dir = dest
			if err := pull.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git pull on %s %v\n", repo, err)
				return
			}
		}
	},
}

func init() {
	repoBranchCmd.AddCommand(repoBranchUpdateCmd)
}

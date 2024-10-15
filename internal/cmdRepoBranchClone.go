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

var branchVersion string

var repoBranchCloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clone Odoo branch repository",
	Long:  `clone Odoo branch repository`,
	Run: func(cmd *cobra.Command, args []string) {
		var version string
		var create bool
		// TODO: implement branchVersion to bypass selection
		// fmt.Println("branchVersion", branchVersion)

		repoShorts, _ := RepoShortCodes("odoo")
		versions := GetCurrentOdooRepos()
		for _, version := range versions {
			repoShorts = removeValue(repoShorts, version)
		}
		versionOptions := []huh.Option[string]{}
		for _, version := range repoShorts {
			versionOptions = append(versionOptions, huh.NewOption(version, version))
		}

		if len(versionOptions) == 0 {
			fmt.Fprintln(os.Stderr, "no more branches to clone")
			return
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Available Odoo Branches").
					Options(versionOptions...).
					Value(&version),

				huh.NewConfirm().
					Title("Clone Branch?").
					Value(&create),
			),
		)
		if err := form.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "odoo version form error %v", err)
			return
		}

		if !create {
			return
		}

		repoDir := viper.GetString("dirs.repo")

		for _, repo := range OdooRepos {
			source := filepath.Join(repoDir, repo)
			dest := filepath.Join(repoDir, version, repo)
			if err := CopyDirectory(source, dest); err != nil {
				fmt.Fprintf(os.Stderr, "copy directory %s to %s failed %v", source, dest, err)
				return
			}

			fetcher := exec.Command("git", "fetch", "origin")
			fetcher.Dir = dest
			if err := fetcher.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git fetch origin on %s %v", repo, err)
				return
			}

			checkout := exec.Command("git", "checkout", version)
			checkout.Dir = dest
			if err := checkout.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git checkout on %s %v", repo, err)
				return
			}

			pull := exec.Command("git", "pull", "origin", version)
			pull.Dir = dest
			if err := pull.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git pull on %s %v", repo, err)
				return
			}
		}
	},
}

func init() {
	repoBranchCmd.AddCommand(repoBranchCloneCmd)
	repoBranchCloneCmd.Flags().StringVarP(&branchVersion, "version", "v", "", "Odoo version")
}

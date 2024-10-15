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
	"github.com/spf13/viper"
)

var repoBaseUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update Odoo source repository",
	Long:  `update Odoo source repository`,
	Run: func(cmd *cobra.Command, args []string) {
		repoDir := viper.GetString("dirs.repo")
		for _, repo := range OdooRepos {
			fmt.Fprintln(os.Stderr, "Updating", repo)

			dest := filepath.Join(repoDir, repo)

			repoHeadShortCode, err := RepoHeadShortCode(repo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "repoHeadShortCode %v\n", err)
				return
			}

			fetch := exec.Command("git", "fetch", "origin")
			fetch.Dir = dest
			if err := fetch.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git fetch origin on %s %v\n", repo, err)
				return
			}

			checkout := exec.Command("git", "checkout", repoHeadShortCode)
			checkout.Dir = dest
			if err := checkout.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git checkout on %s %v\n", repo, err)
				return
			}

			pull := exec.Command("git", "pull", "origin", repoHeadShortCode)
			pull.Dir = dest
			if err := pull.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git pull on %s %v", repo, err)
				return
			}
		}
	},
}

func init() {
	repoBaseCmd.AddCommand(repoBaseUpdateCmd)
}

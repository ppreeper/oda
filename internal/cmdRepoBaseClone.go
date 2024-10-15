/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var repoBaseCloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clone Odoo source repository",
	Long:  `clone Odoo source repository`,
	Run: func(cmd *cobra.Command, args []string) {
		repoDir := viper.GetString("dirs.repo")
		for _, repo := range OdooRepos {
			fmt.Fprintln(os.Stderr, "Cloning", repo)
			username, token := GetGitHubUsernameToken()
			err := CloneUrlDir(
				OdooRepoBase+repo,
				repoDir, repo,
				username, token,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "odoo repo %s clone err %v\n", repo, err)
			}
		}
	},
}

func init() {
	repoBaseCmd.AddCommand(repoBaseCloneCmd)
}

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

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Follow the logs",
	Long:  `Follow the logs`,
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()
		podCmd := exec.Command("incus",
			"exec", project, "-t", "--",
			"journalctl", "-f",
		)
		podCmd.Stdin = os.Stdin
		podCmd.Stdout = os.Stdout
		podCmd.Stderr = os.Stderr
		if err := podCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error getting logs %w", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}

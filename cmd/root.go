package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
}

var rootCmd = &cobra.Command{
	Use:   "oda",
	Short: "Oda is an Odoo administration tool",
	Long:  `A Fast and Flexible Odoo administration tool`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup the odoo project",
	Long:  "backup the odoo project",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires backup dump filename")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("backup command executed", args[0])
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "restore the odoo project",
	Long:  "restore the odoo project",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires backup dump filename")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("restore command executed", args[0])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseConfig(param string) (string, error) {
	// Get variable from odoo config
	file, err := os.Open("./conf/odoo.conf")
	if err != nil {
		return "cannot find file, make sure you are in the odoo project folder", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	vv := []string{}
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), param) {
			vv = strings.Split(scanner.Text(), "=")
			for i := range vv {
				vv[i] = strings.TrimSpace(vv[i])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "scanner error", err
	}
	return vv[1], nil
}

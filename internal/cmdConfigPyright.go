/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pyrightCmd = &cobra.Command{
	Use:   "pyright",
	Short: "Setup pyright settings",
	Long:  `Setup pyright settings`,
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not get current working directory: %w", err)
			return
		}

		version := fmt.Sprintf("%0.1f", viper.GetFloat64("version"))
		dirRepo := viper.GetString("dirs.repo")

		odoo := filepath.Join(dirRepo, version, "odoo")
		enterprise := filepath.Join(dirRepo, version, "enterprise")
		designThemes := filepath.Join(dirRepo, version, "design-themes")
		industry := filepath.Join(dirRepo, version, "industry")

		cfg := map[string]any{}
		cfg["venvPath"] = "."
		cfg["venv"] = ".direnv"
		cfg["executionEnvironments"] = []map[string]any{
			{
				"root": ".",
				"extraPaths": []string{
					odoo,
					enterprise,
					designThemes,
					industry,
					"addons",
				},
			},
		}
		jsonCfg, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not marshal pyright configuration: %w", err)
			return
		}

		pyrightconfig, err := os.Create(filepath.Join(cwd, "pyrightconfig.json"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not create pyrightconfig.json: %w", err)
			return
		}
		defer pyrightconfig.Close()
		pyrightconfig.WriteString(string(jsonCfg))
	},
}

func init() {
	configCmd.AddCommand(pyrightCmd)
}

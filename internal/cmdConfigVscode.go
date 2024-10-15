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

var vscodeCmd = &cobra.Command{
	Use:   "vscode",
	Short: "Setup vscode settings and launch json files",
	Long:  `Setup vscode settings and launch json files`,
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

		if _, err := os.Stat(odoo); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "odoo version does not exist")
			return
		}
		if _, err := os.Stat(filepath.Join(cwd, ".vscode")); os.IsNotExist(err) {
			os.MkdirAll(filepath.Join(cwd, ".vscode"), 0o755)
		}

		// launch.json
		launchCfg := map[string]any{}
		launchCfg["version"] = "0.2.0"
		launchCfg["configurations"] = []map[string]any{
			{
				"name":        "Launch",
				"type":        "python",
				"request":     "launch",
				"stopOnEntry": false,
				"python":      "${command:python.interpreterPath}",
				"program":     "${workspaceRoot}/odoo/odoo-bin",
				"args":        []string{"-c", "${workspaceRoot}/conf/odoo.conf", "-p", "$ODOO_PORT"},
				"cwd":         "${workspaceRoot}",
				"env":         map[string]any{},
				"envFile":     "${workspaceFolder}/.env",
				"console":     "integratedTerminal",
			},
		}
		launchJSON, err := json.MarshalIndent(launchCfg, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not marshal launch configuration: %w", err)
			return
		}

		launch, err := os.Create(filepath.Join(cwd, ".vscode", "launch.json"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not create launch.json: %w", err)
			return
		}
		defer launch.Close()
		launch.WriteString(string(launchJSON))

		// settings.json
		settingsCfg := map[string]any{}
		settingsCfg["python.terminal.executeInFileDir"] = true
		settingsCfg["python.analysis.extraPaths"] = []string{
			odoo,
			enterprise,
			designThemes,
			industry,
			"addons",
		}
		settingsJSON, err := json.MarshalIndent(settingsCfg, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not marshal settings configuration: %w", err)
			return
		}

		settings, err := os.Create(filepath.Join(cwd, ".vscode", "settings.json"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not create settings.json: %w", err)
			return
		}
		defer settings.Close()
		settings.WriteString(string(settingsJSON))
	},
}

func init() {
	configCmd.AddCommand(vscodeCmd)
}

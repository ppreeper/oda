package oda

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ConfigVSCode() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}
	version := GetVersion()
	dirList := GetDirs()
	odoo := filepath.Join(dirList.Repo, version, "odoo")
	enterprise := filepath.Join(dirList.Repo, version, "enterprise")

	if _, err := os.Stat(odoo); os.IsNotExist(err) {
		fmt.Println("odoo version does not exist")
		return nil
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
		return fmt.Errorf("could not marshal launch configuration: %w", err)
	}

	launch, err := os.Create(filepath.Join(cwd, ".vscode", "launch.json"))
	if err != nil {
		return fmt.Errorf("could not create launch.json: %w", err)
	}
	defer launch.Close()
	launch.WriteString(string(launchJSON))

	// settings.json
	settingsCfg := map[string]any{}
	settingsCfg["python.terminal.executeInFileDir"] = true
	settingsCfg["python.analysis.extraPaths"] = []string{
		odoo,
		enterprise,
		"addons",
	}
	settingsJSON, err := json.MarshalIndent(settingsCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal settings configuration: %w", err)
	}

	settings, err := os.Create(filepath.Join(cwd, ".vscode", "settings.json"))
	if err != nil {
		return fmt.Errorf("could not create settings.json: %w", err)
	}
	defer settings.Close()
	settings.WriteString(string(settingsJSON))
	return nil
}

func ConfigPyright() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}
	version := GetVersion()
	dirList := GetDirs()
	odoo := filepath.Join(dirList.Repo, version, "odoo")
	enterrprise := filepath.Join(dirList.Repo, version, "enterprise")

	cfg := map[string]any{}
	cfg["venvPath"] = "."
	cfg["venv"] = ".direnv"
	cfg["executionEnvironments"] = []map[string]any{
		{
			"root": ".",
			"extraPaths": []string{
				odoo,
				enterrprise,
				"addons",
			},
		},
	}
	jsonCfg, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal pyright configuration: %w", err)
	}

	pyrightconfig, err := os.Create(filepath.Join(cwd, "pyrightconfig.json"))
	if err != nil {
		return fmt.Errorf("could not create pyrightconfig.json: %w", err)
	}
	defer pyrightconfig.Close()
	pyrightconfig.WriteString(string(jsonCfg))
	return nil
}

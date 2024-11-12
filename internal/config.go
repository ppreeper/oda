package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/str"
	"gopkg.in/yaml.v3"
)

func (o *ODA) ConfigPyright() error {
	if !IsProject() {
		return nil
	}
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return err
	}
	projectConf, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	version := projectConf.Version
	dirRepo := odaConf.Dirs.Repo

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
		return fmt.Errorf("could not marshal pyright configuration: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}

	pyrightconfig, err := os.Create(filepath.Join(cwd, "pyrightconfig.json"))
	if err != nil {
		return fmt.Errorf("could not create pyrightconfig.json: %w", err)
	}
	defer pyrightconfig.Close()
	pyrightconfig.WriteString(string(jsonCfg))
	return nil
}

func (o *ODA) ConfigVSCode() error {
	// fmt.Println("ConfigVSCode")
	if !IsProject() {
		return nil
	}
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return err
	}
	projectConf, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}

	version := projectConf.Version
	dirRepo := odaConf.Dirs.Repo

	odoo := filepath.Join(dirRepo, version, "odoo")
	enterprise := filepath.Join(dirRepo, version, "enterprise")
	designThemes := filepath.Join(dirRepo, version, "design-themes")
	industry := filepath.Join(dirRepo, version, "industry")

	if _, err := os.Stat(odoo); os.IsNotExist(err) {
		return fmt.Errorf("odoo version does not exist")
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
		designThemes,
		industry,
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

func (o *ODA) HostsUpdate(domain string) error {
	sudouser, _ := os.LookupEnv("SUDO_USER")
	if sudouser == "" {
		fmt.Fprintln(os.Stderr, "not allowed: this requires root access")
		return nil
	}

	odaConf, err := config.LoadOdaConfigUser(sudouser)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading oda config", err)
		return nil
	}
	if domain == "" {
		domain = odaConf.System.Domain
	}
	inc := incus.NewIncus(odaConf)

	hosts, err := os.Open("/etc/hosts")
	if err != nil {
		fmt.Fprintln(os.Stderr, "hosts file read failed %w", err)
		return nil
	}
	defer hosts.Close()

	hostlines := []string{}
	scanner := bufio.NewScanner(hosts)
	for scanner.Scan() {
		hostlines = append(hostlines, scanner.Text())
	}
	begin := slices.Index(hostlines, "#ODABEGIN")
	end := slices.Index(hostlines, "#ODAEND")

	if begin > end {
		fmt.Fprintln(os.Stderr, "host file out of order, edit /etc/hosts manually")
		return nil
	}

	projects := GetCurrentOdooProjectsUser(sudouser)

	instances, err := inc.GetInstances()
	if err != nil {
		fmt.Fprintln(os.Stderr, "instances list failed %w", err)
		return nil
	}

	projectLines := []string{}

	for _, instance := range instances {
		for _, project := range projects {
			if instance.Name == project {
				projectLines = append(projectLines,
					str.RightLen(instance.IP4, " ", 16)+" "+instance.Name+"."+domain)
			}
		}
	}

	instance, err := inc.GetInstance(odaConf.Database.Host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "instance %s not found %v\n", odaConf.Database.Host, err)
		return nil
	}
	projectLines = append(projectLines,
		str.RightLen(instance.IP4, " ", 16)+" "+odaConf.Database.Host+"."+odaConf.System.Domain)

	newHostlines := []string{}
	if begin == -1 && end == -1 {
		newHostlines = append(newHostlines, hostlines...)
		newHostlines = append(newHostlines, "#ODABEGIN")
		newHostlines = append(newHostlines, projectLines...)
		newHostlines = append(newHostlines, "#ODAEND")
	} else {
		newHostlines = append(newHostlines, hostlines[:begin+1]...)
		newHostlines = append(newHostlines, projectLines...)
		newHostlines = append(newHostlines, hostlines[end:]...)
	}

	fo, err := os.Create("/etc/hosts")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error write /etc/hosts file failed %w", err)
		return nil
	}
	defer fo.Close()

	for _, hostline := range newHostlines {
		fo.WriteString(hostline + "\n")
	}

	return nil
}

func (o *ODA) ConfigInit() error {
	HOME, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// $HOME/.config/oda/oda.yaml
	odaConf := config.NewOdaConfig()

	odaConf.Dirs.Repo = filepath.Join(HOME, "workspace/repos/odoo")
	if _, err := os.Stat(odaConf.Dirs.Repo); os.IsNotExist(err) {
		os.MkdirAll(odaConf.Dirs.Repo, 0o755)
	}

	odaConf.Dirs.Project = filepath.Join(HOME, "workspace/odoo")
	if _, err := os.Stat(odaConf.Dirs.Project); os.IsNotExist(err) {
		os.MkdirAll(odaConf.Dirs.Project, 0o755)
	}

	yamlOdaConfOut, err := yaml.Marshal(odaConf)
	if err != nil {
		return fmt.Errorf("yaml marshalling error %w", err)
	}
	fmt.Println(string(yamlOdaConfOut))

	CFGDIR, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	odaFile := filepath.Join(CFGDIR, "oda", "oda.yaml")
	if _, err := os.Stat(odaFile); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(CFGDIR, "oda"), 0o750)
		os.WriteFile(odaFile, yamlOdaConfOut, 0o640)
	}

	return nil
}

package oda

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func AdminInit() error {
	HOME, _ := os.UserHomeDir()

	conf := *NewConf()

	// Repo
	REPO := filepath.Join(HOME, "workspace/repos/odoo")
	if _, err := os.Stat(REPO); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(REPO), 0o755)
	}
	conf.Repo = REPO

	// Project
	PROJECT := filepath.Join(HOME, "workspace/odoo")
	if _, err := os.Stat(filepath.Join(PROJECT, "backups")); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(PROJECT, "backups"), 0o755)
	}
	conf.Project = PROJECT

	confRec, _ := json.Marshal(conf)
	var confMap map[string]any
	json.Unmarshal(confRec, &confMap)

	keys := []string{}
	for field := range confMap {
		keys = append(keys, field)
	}
	sort.Strings(keys)

	// ODA Config
	ODAConfig(keys, confMap)

	return nil
}

func ODAConfig(keys []string, confMap map[string]any) error {
	CONFIG, _ := os.UserConfigDir()
	// ODA Config
	ODOOCONF := filepath.Join(CONFIG, "oda", "oda.conf")
	if _, err := os.Stat(ODOOCONF); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(CONFIG, "oda"), 0o755)
		f, err := os.Create(ODOOCONF)
		if err != nil {
			return err
		}
		defer f.Close()
		for _, field := range keys {
			f.WriteString(fmt.Sprintf("%s=%s\n", field, confMap[field]))
		}
	}
	return nil
}

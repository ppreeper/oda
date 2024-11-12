package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ppreeper/oda/lib"
	"gopkg.in/yaml.v3"
)

func LoadOdaConfigUser(user string) (*OdaConf, error) {
	var odaConf *OdaConf

	cfgDir := lib.UserConfigDir(user)

	yamlFilename := filepath.Join(cfgDir, "oda", "oda.yaml")

	yamlFile, err := os.ReadFile(yamlFilename)
	if err != nil {
		return nil, fmt.Errorf("oda.yaml not found in %s", cfgDir)
	}

	err = yaml.Unmarshal(yamlFile, &odaConf)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	return odaConf, nil
}

func LoadOdaConfig() (*OdaConf, error) {
	var odaConf *OdaConf
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user config dir: %w", err)
	}

	yamlFilename := filepath.Join(cfgDir, "oda", "oda.yaml")

	yamlFile, err := os.ReadFile(yamlFilename)
	if err != nil {
		return nil, fmt.Errorf("oda.yaml not found in %s", cfgDir)
	}

	err = yaml.Unmarshal(yamlFile, &odaConf)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	return odaConf, nil
}

func NewOdaConfig() *OdaConf {
	return &OdaConf{
		Database: OdaDatabase{
			Host:     "db",
			Port:     5432,
			Version:  17,
			Image:    "debian/12",
			Username: "odoo",
			Password: "odooodoo",
		},
		Dirs: OdaDirs{
			Repo:    "/home/odoo/workspace/repos/odoo",
			Project: "/home/odoo/workspace/odoo",
		},
		Incus: OdaIncus{
			Socket:      "/var/lib/incus/unix.socket",
			Type:        "unix",
			URL:         "http://unix.socket/1.0",
			LimitCPU:    2,
			LimitMemory: "2GiB",
		},
		System: OdaSystem{
			Domain: "local",
			SSHKey: "id_rsa",
		},
	}
}

type OdaDatabase struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Version  int    `json:"version" yaml:"version"`
	Image    string `json:"image" yaml:"image"`
}
type OdaDirs struct {
	Repo    string `json:"repo" yaml:"repo"`
	Project string `json:"project" yaml:"project"`
}
type OdaIncus struct {
	Socket      string `json:"socket" yaml:"socket"`
	Type        string `json:"type" yaml:"type"`
	URL         string `json:"url" yaml:"url"`
	LimitCPU    int    `json:"limit_cpu" yaml:"limit_cpu"`
	LimitMemory string `json:"limit_memory" yaml:"limit_memory"`
}
type OdaSystem struct {
	Domain string `json:"domain" yaml:"domain"`
	SSHKey string `json:"ssh_key" yaml:"ssh_key"`
}
type OdaConf struct {
	Database OdaDatabase `json:"database" yaml:"database"`
	Dirs     OdaDirs     `json:"dirs" yaml:"dirs"`
	Incus    OdaIncus    `json:"incus" yaml:"incus"`
	System   OdaSystem   `json:"system" yaml:"system"`
}

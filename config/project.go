package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type OdaProject struct {
	Version string `json:"version"`
}

func LoadProjectConfig() (*OdaProject, error) {
	var config *OdaProject
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get current working directory: %w", err)
	}
	yamlFilename := filepath.Join(cwd, ".oda.yaml")
	yamlFile, err := os.ReadFile(yamlFilename)
	if err != nil {
		return nil, fmt.Errorf("could not read project config file: %w", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}
	return config, nil
}

func (p *OdaProject) WriteConfig(configPath string) error {
	odaYamlData, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, odaYamlData, 0o644); err != nil {
		return fmt.Errorf("cannot create project .oda.yaml file %w", err)
	}
	return nil
}

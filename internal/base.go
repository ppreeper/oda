package internal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
)

func (o *ODA) BaseCreate() error {
	odooInstances := getBaseImages()

	vers := []string{}
	for _, version := range GetCurrentOdooRepos() {
		vers = append(vers, strings.Split(version, ".")[0])
	}
	vers = removeDuplicate(vers)

	versMatch := []string{}
	for _, repo := range vers {
		repoVersion := "odoo-" + repo + "-0"
		for _, odooInstance := range odooInstances {
			if repoVersion == odooInstance {
				versMatch = append(versMatch, repo)
			}
		}
	}
	for _, match := range versMatch {
		vers = removeValue(vers, match)
	}

	if len(vers) == 0 {
		return fmt.Errorf("no versions to create, rebuild if necessary")
	}

	versionOptions := []huh.Option[string]{}
	for _, version := range vers {
		versionOptions = append(versionOptions, huh.NewOption(version, version+".0"))
	}

	var (
		version string
		create  bool
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Base Version").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Create Odoo Base Image?").
				Value(&create),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("create base form error %w", err)
	}
	if create {
		if err := o.BaseCreateScript(version); err != nil {
			return fmt.Errorf("create base %s error %w", version, err)
		}
	}

	return nil
}

func (o *ODA) BaseUpdate() error {
	odooInstances := getBaseImages()

	if len(odooInstances) == 0 {
		return fmt.Errorf("no base images found")
	}

	versionOptions := []huh.Option[string]{}
	for _, version := range odooInstances {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	var (
		version string
		update  bool
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Base Version").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Update Odoo base image packages?").
				Value(&update),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("updating base form error %w", err)
	}
	if !update {
		return nil
	}

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	inc.SetInstanceState(version, "start")
	roleUpdate(version)
	inc.SetInstanceState(version, "stop")

	return nil
}

func (o *ODA) BaseDestroy() error {
	odooInstances := getBaseImages()

	if len(odooInstances) == 0 {
		return fmt.Errorf("no base images found")
	}

	versionOptions := []huh.Option[string]{}
	for _, version := range odooInstances {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	var (
		version string
		destroy bool
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Base Version").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Destroy Odoo Base Image?").
				Value(&destroy),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("destroy base form error %w", err)
	}

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if destroy {
		inc.SetInstanceState(version, "stop")
		inc.DeleteInstance(version)
	}

	return nil
}

func getBaseImages() []string {
	versions := GetCurrentOdooRepos()
	var odooVersions []string
	for _, version := range versions {
		odooVersions = append(odooVersions, "odoo-"+strings.ReplaceAll(version, ".", "-"))
	}
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)
	instances, err := inc.GetInstances()
	if err != nil {
		fmt.Println("getBaseImages error", err)
	}
	var odooInstances []string
	for _, inst := range instances {
		for _, version := range odooVersions {
			if inst.Name == version {
				odooInstances = append(odooInstances, inst.Name)
			}
		}
	}
	return odooInstances
}

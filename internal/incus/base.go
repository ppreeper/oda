package oda

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

func BaseCreatePrompt() error {
	var (
		version string
		create  bool
	)

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

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
	if err := BaseCreate(version); err != nil {
		return fmt.Errorf("create base %s error %w", version, err)
	}
	return nil
}

func BaseDestroyPrompt() error {
	var (
		version string
		destroy bool
	)

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

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
	if err := BaseDestroy(version); err != nil {
		return fmt.Errorf("destroy base %s error %w", version, err)
	}
	return nil
}

func BaseRebuildPrompt() error {
	var (
		version string
		rebuild bool
	)

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Base Version").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Rebuild Odoo Base Image?").
				Value(&rebuild),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("rebuild base form error %w", err)
	}
	if err := BaseDestroy(version); err != nil {
		return fmt.Errorf("destroy base %s error %w", version, err)
	}
	if err := BaseCreate(version); err != nil {
		return fmt.Errorf("create base %s error %w", version, err)
	}
	return nil
}

func BaseUpdatePrompt() error {
	var (
		version string
		update  bool
	)

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

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
	vers := "odoo-" + strings.Replace(version, ".", "-", -1)
	instance, err := GetInstance(vers)
	if err != nil {
		return fmt.Errorf("get instance %s error %w", vers, err)
	}
	switch instance.State {
	case "STOPPED":
		if err := IncusStart(vers); err != nil {
			return fmt.Errorf("start instance %s error %w", vers, err)
		}
		if err := roleUpdate(vers); err != nil {
			return fmt.Errorf("update instance %s error %w", vers, err)
		}
		if err := IncusStop(vers); err != nil {
			return fmt.Errorf("stop instance %s error %w", vers, err)
		}
	case "RUNNING":
		if err := roleUpdate(vers); err != nil {
			return fmt.Errorf("update instance %s error %w", vers, err)
		}
	}
	return nil
}

////////////////////////

func BaseCreate(version string) error {
	conf := GetConf()
	vers := "odoo-" + strings.Replace(version, ".", "-", -1)

	if err := IncusLaunch(vers, conf.OSImage); err != nil {
		return fmt.Errorf("launch instance %s %w", vers, err)
	}

	if err := WaitForInstance(vers); err != nil {
		return fmt.Errorf("wait for instance %s failed %w", vers, err)
	}

	if err := roleUpdateScript(vers); err != nil {
		return fmt.Errorf("roleUpdateScript %s failed %w", vers, err)
	}

	if err := roleUpdate(vers); err != nil {
		return fmt.Errorf("roleUpdate %s failed %w", vers, err)
	}

	if err := roleBaseline(vers); err != nil {
		return fmt.Errorf("roleBaseline %s failed %w", vers, err)
	}

	if err := roleOdooUser(vers); err != nil {
		return fmt.Errorf("roleOdooUser %s failed %w", vers, err)
	}

	if err := roleOdooDirs(vers); err != nil {
		return fmt.Errorf("roleOdooDirs %s failed %w", vers, err)
	}

	if err := rolePostgresqlRepo(vers); err != nil {
		return fmt.Errorf("rolePostgresqlRepo %s failed %w", vers, err)
	}

	if err := rolePostgresqlClient(vers); err != nil {
		return fmt.Errorf("rolePostgresqlClient %s failed %w", vers, err)
	}

	if err := roleWkhtmltopdf(vers); err != nil {
		return fmt.Errorf("roleWkhtmltopdf %s failed %w", vers, err)
	}

	if err := roleOdooBasePackages(vers, version); err != nil {
		return fmt.Errorf("roleOdooBasePackages %s failed %w", vers, err)
	}

	if err := pip3Install(vers, "ebaysdk", "google-auth"); err != nil {
		return fmt.Errorf("pip3Install %s failed %w", vers, err)
	}

	if err := npmInstall(vers, "rtlcss"); err != nil {
		return fmt.Errorf("npmInstall %s failed %w", vers, err)
	}

	if err := roleGeoIP2DB(vers); err != nil {
		return fmt.Errorf("roleGeoIP2DB %s failed %w", vers, err)
	}

	if err := rolePaperSize(vers); err != nil {
		return fmt.Errorf("rolePaperSize %s failed %w", vers, err)
	}

	if err := roleOdooNode(vers); err != nil {
		return fmt.Errorf("roleOdooNode %s failed %w", vers, err)
	}

	if err := roleOdooService(vers); err != nil {
		return fmt.Errorf("roleOdooService %s failed %w", vers, err)
	}

	if err := roleCaddy(vers); err != nil {
		return fmt.Errorf("roleCaddy %s failed %w", vers, err)
	}

	if err := roleCaddyService(vers); err != nil {
		return fmt.Errorf("roleCaddyService %s failed %w", vers, err)
	}

	if err := IncusStop(vers); err != nil {
		return fmt.Errorf("IncusStop %s failed %w", vers, err)
	}

	return nil
}

func BaseDestroy(version string) error {
	vers := "odoo-" + strings.Replace(version, ".", "-", -1)
	if err := IncusDelete(vers); err != nil {
		return fmt.Errorf("destroy instance %s failed %w", vers, err)
	}
	fmt.Println("destroying:", vers)
	return nil
}

func aptInstall(instanceName string, pkgs ...string) error {
	pkg := []string{"apt-get", "install", "-y", "--no-install-recommends"}
	pkg = append(pkg, pkgs...)
	if err := IncusExec(instanceName, pkg...); err != nil {
		return fmt.Errorf("apt-get install failed %w", err)
	}
	return nil
}

func pip3Install(name string, pkgs ...string) error {
	fmt.Println("pip3Install:", name)
	pkg := []string{"pip3", "install"}
	pkg = append(pkg, pkgs...)
	if err := IncusExec(name, pkg...); err != nil {
		return fmt.Errorf("pip3 install failed %w", err)
	}
	return nil
}

func npmInstall(instanceName string, pkgs ...string) error {
	fmt.Println("npmInstall:", instanceName)
	pkg := []string{"npm", "install", "-g"}
	pkg = append(pkg, pkgs...)
	if err := IncusExec(instanceName, pkg...); err != nil {
		return fmt.Errorf("npm install failed %w", err)
	}
	return nil
}

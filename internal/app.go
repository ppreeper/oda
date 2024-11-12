package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/oda/ui"
)

func moduleList(modules ...string) string {
	mods := []string{}
	for _, mod := range modules {
		mm := strings.Split(mod, ",")
		mods = append(mods, mm...)
	}
	return strings.Join(removeDuplicate(mods), ",")
}

func (o *ODA) InstanceAppInstallUpgrade(install bool, modules ...string) error {
	if !IsProject() {
		return nil
	}
	iu := "upgrade"
	if install {
		iu = "install"
	}
	_, project := lib.GetProject()
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExecVerbose(project, "odas", iu, moduleList(modules...)); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("error installing/upgrading modules %v"), err)
	}

	return nil
}

func (o *ODA) Scaffold(module string) error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	if err := inc.IncusExecVerbose(project, "odas", "scaffold", module); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("error scaffolding module %v"), err)
	}

	return nil
}

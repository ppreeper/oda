package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/oda/ui"
)

func (o *ODA) Backup() error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	uid, err := inc.IncusGetUid(project, "odoo")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not get odoo uid", err)
	}

	if err := exec.Command("incus", "exec", project, "--user", uid, "-t",
		"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
		"odas", "backup",
	).Run(); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("error backing up project %v"), err)
		return nil
	}

	return nil
}

// GetCurrentOdooProjects Get Current Odoo Projects
func GetOdooBackups(project string) (backups, addons []string) {
	odaConf, _ := config.LoadOdaConfig()

	entries, err := os.ReadDir(filepath.Join(odaConf.Dirs.Project, "backups"))
	if err != nil {
		fmt.Println(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			fname := strings.Split(entry.Name(), "__")
			if len(fname) == 2 {
				backups = append(backups, entry.Name())
			} else if len(fname) == 3 {
				addons = append(addons, entry.Name())
			}
		}
	}
	slices.Sort(backups)
	slices.Sort(addons)
	backups = removeDuplicate(backups)
	addons = removeDuplicate(addons)
	if project != "" {
		backups = selectOnly(backups, project)
		addons = selectOnly(addons, project)
	}
	return
}

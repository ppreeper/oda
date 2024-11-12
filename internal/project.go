package internal

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/oda/ui"
)

// ProjectInit
// build project directory based on prompts
func (o *ODA) ProjectInit() error {
	projects := GetCurrentOdooProjects()
	versions := GetCurrentOdooRepos()

	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	var (
		name    string
		edition string
		version string
		create  bool
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Value(&name).
				Validate(func(str string) error {
					// check if project already exists
					if existsIn(projects, str) {
						return fmt.Errorf("project %s already exists", str)
					}
					if str == "" {
						return fmt.Errorf("project name is required")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Odoo Edition").
				Options(
					huh.NewOption("Community", "community"),
					huh.NewOption("Enterprise", "enterprise").Selected(true),
				).
				Value(&edition),

			huh.NewSelect[string]().
				Title("Odoo Branch").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Create Project?").
				Value(&create),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("project init form error %w", err)
	}
	if !create {
		return nil
	}

	if err := projectSetup(name, edition, version, o.EmbedFS); err != nil {
		return fmt.Errorf("project setup failed %w", err)
	}

	return nil
}

// projectSetup Project Config Setup
func projectSetup(projectName, edition, version string, embedFS embed.FS) error {
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("creating project directory"))
	projectDir := filepath.Join(odaConf.Dirs.Project, projectName)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return fmt.Errorf("cannot create project directory %w", err)
	}

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("creating project subdirectories"))
	for _, pdir := range []string{"addons", "conf", "data"} {
		projectSubDir := filepath.Join(projectDir, pdir)
		if err := os.MkdirAll(projectSubDir, 0o755); err != nil {
			return fmt.Errorf("cannot create project subdirectory %s %w", pdir, err)
		}
	}

	// odoo.conf
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("creating project odoo.conf"))
	odooConf := config.NewOdooConfig()
	odooConf.Write(projectName, projectDir, edition, embedFS)

	// .env (for vscode env injections)
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("creating project .env file"))
	envFile := filepath.Join(projectDir, ".env")
	if err := os.WriteFile(envFile, []byte("ODOO_V="+version), 0o644); err != nil {
		return fmt.Errorf("cannot create project .env file %w", err)
	}

	// .oda.yaml
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("creating project .oda.yaml"))
	projectCfg := &config.OdaProject{}
	projectCfg.Version = version
	err = projectCfg.WriteConfig(filepath.Join(projectDir, ".oda.yaml"))
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, ui.StepStyle.Render("project %s init complete")+"\n", projectName)
	return nil
}

// ####################################
// ProjectReset
// reset project data directory and db
// stop instance
// drop instance db
// remove instance data directory contents
func (o *ODA) ProjectReset() error {
	if !IsProject() {
		return nil
	}
	confim := ui.AreYouSure("reset the project")
	if !confim {
		return fmt.Errorf("reset the project canceled")
	}

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

	// stop
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("stopping the instance"))
	cwd, project := lib.GetProject()
	inc.SetInstanceState(project, "stop")
	// instanceStatus := GetInstanceState(project)
	// fmt.Fprintln(os.Stderr, project, "instanceStatus.Metadata.Status", instanceStatus.Metadata.Status)

	// rm -rf data/*
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("removing data files"))
	if err := RemoveContents(filepath.Join(cwd, "data")); err != nil {
		return fmt.Errorf("data files removal failed %w", err)
	}

	// drop db
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("dropping database"))
	odooConf, _ := config.LoadOdooConfig(cwd)
	dbhost := odooConf.DbHost
	dbname := odooConf.DbName

	uid, err := inc.IncusGetUid(dbhost, "postgres")
	if err != nil {
		return fmt.Errorf("could not get postgres user id: %w", err)
	}
	if err := exec.Command("incus", "exec", dbhost, "--user", uid, "-t", "--",
		"dropdb", "--if-exists", "-U", "postgres", "-f", dbname,
	).Run(); err != nil {
		return fmt.Errorf("could not drop postgresql database %s error: %w", dbname, err)
	}
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("project reset complete"))
	return nil
}

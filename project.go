package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

// projectIinit Project Init
func projectIinit() error {
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
	if err := projectSetup(name, edition, version); err != nil {
		return fmt.Errorf("project setup failed %w", err)
	}
	return nil
}

// projectBranch Project Branch Init
func projectBranch() error {
	projects := GetCurrentOdooProjects()

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	var (
		name    string
		branch  string
		url     string
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
					if existsIn(projects, name+"-"+branch) {
						return fmt.Errorf("project %s already exists", str)
					}
					return nil
				}),
			huh.NewInput().
				Title("Project Branch").
				Value(&branch).
				Validate(func(str string) error {
					// check if project already exists
					if existsIn(projects, name+"-"+branch) {
						return fmt.Errorf("project %s already exists", str)
					}
					return nil
				}),
			huh.NewInput().
				Title("Project URL").
				Value(&url),

			huh.NewSelect[string]().
				Title("Odoo Edition").
				Value(&edition).
				Options(
					huh.NewOption("Community", "community"),
					huh.NewOption("Enterprise", "enterprise").Selected(true),
				),

			huh.NewSelect[string]().
				Title("Odoo Branch").
				Value(&version).
				Options(versionOptions...),

			huh.NewConfirm().
				Title("Create Project Branch?").
				Value(&create),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("project branch form error %w", err)
	}
	project := name + "-" + branch
	if err := projectSetup(project, edition, version); err != nil {
		return fmt.Errorf("project branch setup failed %w", err)
	}
	// clone repo branch
	dirs := GetDirs()
	username, token := getGitHubUsernameToken()
	if err := cloneUrlDir(
		url,
		filepath.Join(dirs.Project, project, "addons"),
		branch,
		username,
		token,
	); err != nil {
		return fmt.Errorf("project branch addons clone failed %w", err)
	}
	return nil
}

// projectSetup Project Config Setup
func projectSetup(projectName, edition, version string) error {
	dirs := GetDirs()
	projectDir := filepath.Join(dirs.Project, projectName)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return fmt.Errorf("cannot create project directory %w", err)
	}
	for _, pdir := range []string{"addons", "conf", "data"} {
		projectSubDir := filepath.Join(projectDir, pdir)
		if err := os.MkdirAll(projectSubDir, 0o755); err != nil {
			return fmt.Errorf("cannot create project subdirectory %s %w", pdir, err)
		}
	}
	// odoo.conf
	odooConfFile := filepath.Join(projectDir, "conf", "odoo.conf")
	writeOdooConf(odooConfFile, projectName, edition)

	// envrc
	envFile := filepath.Join(projectDir, ".env")
	if err := os.WriteFile(envFile, []byte("ODOO_V="+version), 0o644); err != nil {
		return fmt.Errorf("cannot create project env file %w", err)
	}
	return nil
}

// writeOdooConf Write Odoo Configfile
func writeOdooConf(file, projectName, edition string) error {
	t := time.Now()
	conf := GetConf()
	projectName = strings.ReplaceAll(projectName, "-", "_")
	dbname := projectName + "_" + t.Format("20060102150405")
	enterprise_dir := ""
	if edition == "enterprise" {
		enterprise_dir = "/opt/odoo/enterprise,"
	}
	fo, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fo.Close()
	fo.WriteString("[options]" + "\n")
	fo.WriteString("addons_path = /opt/odoo/odoo/addons," + enterprise_dir + "/opt/odoo/addons" + "\n")
	fo.WriteString("data_dir = /opt/odoo/data" + "\n")
	fo.WriteString("admin_passwd = adminadmin" + "\n")
	fo.WriteString("without_demo = all" + "\n")
	fo.WriteString("csv_internal_sep = ;" + "\n")
	fo.WriteString("reportgz = false" + "\n")
	fo.WriteString("server_wide_modules = base,web" + "\n")
	fo.WriteString("db_host = " + conf.DBHost + "\n")
	fo.WriteString("db_port = " + conf.DBPort + "\n")
	fo.WriteString("db_maxconn = 8" + "\n")
	fo.WriteString("db_user = " + conf.DBUsername + "\n")
	fo.WriteString("db_password = " + conf.DBUserpass + "\n")
	fo.WriteString("db_name = " + dbname + "\n")
	fo.WriteString("db_template = template0" + "\n")
	fo.WriteString("db_sslmode = disable" + "\n")
	fo.WriteString("list_db = false" + "\n")
	fo.WriteString("proxy = true" + "\n")
	fo.WriteString("proxy_mode = true" + "\n")
	fo.WriteString("logfile = /dev/stderr" + "\n")
	fo.WriteString("log_level = debug" + "\n")
	fo.WriteString("log_handler = odoo.tools.convert:DEBUG" + "\n")
	fo.WriteString("workers = 0" + "\n")
	return nil
}

// projectRebuild Rebuild project with db and filestore of another project but with current addons
func projectRebuild() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}

	projects := GetCurrentOdooProjects()
	projectOptions := []huh.Option[string]{}
	for _, project := range projects {
		projectOptions = append(projectOptions, huh.NewOption(project, project))
	}
	var sourceProject string
	var create bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Available Odoo Projects").
				Options(projectOptions...).
				Value(&sourceProject),

			huh.NewConfirm().
				Title("Rebuild Project?").
				Value(&create),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("odoo version form error %w", err)
	}

	if !create {
		return fmt.Errorf("rebuild project canceled")
	}

	// stop target
	instanceStop()

	// remove targets files, copy from source to target
	dirs := GetDirs()
	cwd, _ := GetProject()
	dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	destdB := GetOdooConf(cwd, "db_name")
	dbpassword := GetOdooConf(cwd, "db_password")
	sourceDB := GetOdooConf(filepath.Join(dirs.Project, sourceProject), "db_name")

	sourceData := filepath.Join(dirs.Project, sourceProject)
	destData := filepath.Join(cwd, "data")

	if err := RemoveContents(destData); err != nil {
		return fmt.Errorf("data files removal failed %w", err)
	}

	sourceFilestore := filepath.Join(sourceData, "data", "filestore", sourceDB)
	destFilestore := filepath.Join(destData, "filestore", destdB)
	if err := os.MkdirAll(destFilestore, 0o755); err != nil {
		return fmt.Errorf("cannot create data subdirectory %s %w", destFilestore, err)
	}

	if err := CopyDirectory(sourceFilestore, destFilestore); err != nil {
		return fmt.Errorf("copy directory %s to %s failed %w", sourceData, destData, err)
	}

	// clone from target database
	DBClone(dbhost, sourceDB, destdB, dbuser, dbpassword)
	DBReset(dbhost, destdB, dbuser, dbpassword)
	return nil
}

// projectReset clear the data directory and drop database"""
func projectReset() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	confim := AreYouSure("reset the project")
	if !confim {
		return fmt.Errorf("reset the project canceled")
	}
	// stop
	if err := instanceStop(); err != nil {
		return err
	}
	// rm -rf data/*
	cwd, _ := GetProject()
	if err := RemoveContents(filepath.Join(cwd, "data")); err != nil {
		return fmt.Errorf("data files removal failed %w", err)
	}
	// drop db
	dbhost := GetOdooConf(cwd, "db_host")
	dbname := GetOdooConf(cwd, "db_name")
	podCmd := exec.Command("podman",
		"exec", "-it", dbhost, "dropdb", "-U", "postgres", dbname)
	if err := podCmd.Run(); err != nil {
		return fmt.Errorf("database drop failed %w", err)
	}
	return nil
}

// projectHostsFile Update /etc/hosts file with projectname
func projectHostsFile() error {
	sudouser, _ := os.LookupEnv("SUDO_USER")
	if sudouser == "" {
		return fmt.Errorf("not allowed: this requires root access")
	}
	hosts, err := os.Open("/etc/hosts")
	if err != nil {
		return fmt.Errorf("hosts file read failed %w", err)
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
		return fmt.Errorf("host file out of order, edit /etc/hosts manually")
	}

	projects := GetCurrentOdooProjects()
	conf := GetConf()
	brAddr := conf.BRAddr

	projectLines := []string{}
	projectLines = append(projectLines, brAddr+" "+conf.DBHost)
	for _, project := range projects {
		projectLines = append(projectLines, brAddr+" "+project+".local")
	}

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
		return err
	}
	defer fo.Close()
	for _, hostline := range newHostlines {
		fo.WriteString(hostline + "\n")
	}
	return nil
}

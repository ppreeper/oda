package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
)

func adminBackup() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()

	if err := exec.Command("podman",
		"exec", "-it", project+".local", "oda_db.py", "-b",
	).Run(); err != nil {
		return fmt.Errorf("backup failed %w", err)
	}
	return nil
}

// adminRestore Restore from backup file
func adminRestore() error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}

	backups, addons := GetOdooBackups()

	backupOptions := []huh.Option[string]{}
	for _, backup := range backups {
		backupOptions = append(backupOptions, huh.NewOption(backup, backup))
	}
	addonOptions := []huh.Option[string]{}
	addonOptions = append(addonOptions, huh.NewOption("None", "none"))
	for _, addon := range addons {
		addonOptions = append(addonOptions, huh.NewOption(addon, addon))
	}

	var (
		backupFile string
		addonFile  string
		confirm    bool
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Odoo Backup File").
				Options(backupOptions...).
				Value(&backupFile),

			huh.NewSelect[string]().
				Title("Odoo Addon File").
				Options(addonOptions...).
				Value(&addonFile),

			huh.NewConfirm().
				Title("Restore Project?").
				Value(&confirm),
		),
	)
	if err := form.Run(); err != nil {
		fmt.Println("form error", err)
	}

	fmt.Println("restore from backup file " + backupFile)
	if err := RestoreDBTar(backupFile, true); err != nil {
		return err
	}

	if addonFile != "none" {
		fmt.Println("restore from addon file " + addonFile)
		if err := RestoreAddonsTar(addonFile); err != nil {
			return err
		}
	}

	return nil
}

// RestoreDBTar Restore Odoo DB from backup
func RestoreDBTar(backupFile string, copyDB bool) error {
	cwd, _ := GetProject()
	dirs := GetDirs()
	source := filepath.Join(dirs.Project, "backups", backupFile)

	dbname := GetOdooConf(cwd, "db_name")
	dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	dbtemplate := GetOdooConf(cwd, "db_template")

	// drop target database
	if err := exec.Command("podman",
		"exec", "-it", dbhost,
		"dropdb", "--if-exists", "-U", "postgres", "-f", dbname,
	).Run(); err != nil {
		return fmt.Errorf("could not drop postgresql database %s error: %w", dbname, err)
	}

	// create new postgresql database
	if err := exec.Command("podman",
		"exec", "-it", dbhost,
		"createdb", "-U", "postgres",
		"--encoding", "unicode",
		"--lc-collate", "C",
		"-T", dbtemplate,
		"-O", dbuser, dbname,
	).Run(); err != nil {
		return fmt.Errorf("could not create postgresql database %s error: %w", dbname, err)
	}

	// restore postgresql database
	tarpgCmd := exec.Command("tar",
		"Oaxf", source, "./dump.sql",
	)
	pgCmd := exec.Command("psql",
		"-h", dbhost, "-U", dbuser, "--dbname", dbname, "-q",
	)
	r, w := io.Pipe()
	tarpgCmd.Stdout = w
	pgCmd.Stdin = r

	tarpgCmd.Start()
	pgCmd.Start()
	tarpgCmd.Wait()
	w.Close()
	pgCmd.Wait()

	// restore data filestore
	data := filepath.Join(cwd, "data")
	if err := RemoveContents(data); err != nil {
		return fmt.Errorf("data files removal failed %w", err)
	}
	filestore := filepath.Join(data, "filestore", dbname)
	if err := os.MkdirAll(filestore, 0o755); err != nil {
		return fmt.Errorf("filestore directory creation failed %w", err)
	}
	tarCmd := exec.Command("tar",
		"axf", source, "-C", filestore, "--strip-components=2", "./filestore",
	)
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("filestore restore failed %w", err)
	}

	if copyDB {
		if err := DBReset(dbhost, dbname, dbuser); err != nil {
			return err
		}
	}
	return nil
}

// DBReset Database Reset DBUUID and remove MCode
func DBReset(dbhost, dbname, dbuser string) error {
	sqlq := []string{}
	sqlq = append(sqlq, "delete from ir_config_parameter where key='database.enterprise_code'")
	sqlq = append(sqlq, `update ir_config_parameter
		set value=(select gen_random_uuid())
		where key = 'database.uuid'`,
	)
	sqlq = append(sqlq,
		`insert into ir_config_parameter
	    (key,value,create_uid,create_date,write_uid,write_date) values
	    ('database.expiration_date',(current_date+'3 months'::interval)::timestamp,1,
	    current_timestamp,1,current_timestamp)
	    on conflict (key)
	    do update set value = (current_date+'3 months'::interval)::timestamp;`,
	)
	for _, q := range sqlq {
		podCmd := exec.Command("podman",
			"exec", "-it", dbhost, "psql", "-U", dbuser, "-c", q, dbname,
		)
		if err := podCmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// RestoreAddonsTar Restore Odoo DB addons folders
func RestoreAddonsTar(addonsFile string) error {
	cwd, _ := GetProject()
	dirs := GetDirs()
	source := filepath.Join(dirs.Project, "backups", addonsFile)
	dest := filepath.Join(cwd, "addons")
	if err := RemoveContents(dest); err != nil {
		return err
	}
	podCmd := exec.Command("tar",
		"axf", source, "-C", dest, ".",
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		return err
	}
	return nil
}

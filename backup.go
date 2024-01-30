package oda

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
)

func AdminBackup() error {
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

func AdminRestoreNode(move bool) error {
	backups, addons := GetOdooBackupsNode()

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
	if err := restoreDBTarNode(backupFile, move); err != nil {
		return err
	}

	if addonFile != "none" {
		fmt.Println("restore from addon file " + addonFile)
		if err := restoreAddonsTarNode(addonFile); err != nil {
			return err
		}
	}

	return nil
}

// AdminRestore Restore from backup file
func AdminRestore(move bool) error {
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
	if err := restoreDBTar(backupFile, move); err != nil {
		return err
	}

	if addonFile != "none" {
		fmt.Println("restore from addon file " + addonFile)
		if err := restoreAddonsTar(addonFile); err != nil {
			return err
		}
	}

	return nil
}

func restoreDBTarNode(backupFile string, moveDB bool) error {
	odir := "/opt/odoo"
	source := filepath.Join(odir, "backups", backupFile)

	dbname := GetOdooConf(odir, "db_name")
	dbhost := GetOdooConf(odir, "db_host")
	dbuser := GetOdooConf(odir, "db_user")
	dbpassword := GetOdooConf(odir, "db_password")
	dbtemplate := GetOdooConf(odir, "db_template")

	fmt.Println("sudo", "-u", "postgres",
		"dropdb", "--if-exists", "-f", dbname)

	// drop target database
	if err := exec.Command("sudo", "-u", "postgres",
		"dropdb", "--if-exists", "-f", dbname).Run(); err != nil {
		return fmt.Errorf("could not drop postgresql database %s error: %w", dbname, err)
	}

	fmt.Println("sudo", "-u", "postgres",
		"createdb",
		"--encoding", "unicode",
		"--lc-collate", "C",
		"-T", dbtemplate,
		"-O", dbuser, dbname)
	// create new postgresql database
	if err := exec.Command("sudo", "-u", "postgres",
		"createdb",
		"--encoding", "unicode",
		"--lc-collate", "C",
		"-T", dbtemplate,
		"-O", dbuser, dbname,
	).Run(); err != nil {
		return fmt.Errorf("could not create postgresql database %s error: %w", dbname, err)
	}

	fmt.Println("tar",
		"Oaxf", source, "./dump.sql")
	// restore postgresql database
	tarpgCmd := exec.Command("tar",
		"Oaxf", source, "./dump.sql",
	)
	fmt.Println("PGPASSWORD=" + dbpassword)
	fmt.Println("psql",
		"-h", dbhost, "-U", dbuser, "--dbname", dbname, "-q")
	pgCmd := exec.Command("psql",
		"-h", dbhost, "-U", dbuser, "--dbname", dbname, "-q",
	)
	pgCmd.Env = append(pgCmd.Env, "PGPASSWORD="+dbpassword)

	r, w := io.Pipe()
	tarpgCmd.Stdout = w
	pgCmd.Stdin = r

	tarpgCmd.Start()
	pgCmd.Start()
	tarpgCmd.Wait()
	w.Close()
	pgCmd.Wait()

	// restore data filestore
	data := filepath.Join(odir, "data")
	fmt.Println("data", data)
	if err := RemoveContents(data); err != nil {
		return fmt.Errorf("data files removal failed %w", err)
	}
	filestore := filepath.Join(data, "filestore", dbname)
	fmt.Println("filestore", filestore)
	if err := os.MkdirAll(filestore, 0o755); err != nil {
		return fmt.Errorf("filestore directory creation failed %w", err)
	}
	fmt.Println("tar",
		"axf", source, "-C", filestore, "--strip-components=2", "./filestore")
	tarCmd := exec.Command("tar",
		"axf", source, "-C", filestore, "--strip-components=2", "./filestore",
	)
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("filestore restore failed %w", err)
	}

	// if not moveDB then reset DBUUID and remove MCode
	if !moveDB {
		fmt.Println("copy the db")
		// if err := dbReset(dbhost, dbname, dbuser, dbpassword); err != nil {
		// 	return err
		// }
	}
	return nil
}

// restoreDBTar Restore Odoo DB from backup
func restoreDBTar(backupFile string, moveDB bool) error {
	cwd, _ := GetProject()
	dirs := GetDirs()
	source := filepath.Join(dirs.Project, "backups", backupFile)

	dbname := GetOdooConf(cwd, "db_name")
	dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
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
	pgCmd.Env = append(pgCmd.Env, "PGPASSWORD="+dbpassword)

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

	// if not moveDB then reset DBUUID and remove MCode
	if !moveDB {
		if err := dbReset(dbhost, dbname, dbuser, dbpassword); err != nil {
			return err
		}
	}
	return nil
}

func dbClone(dbhost, sourceDB, destDB, dbuser, dbpassword string) error {
	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Username: dbuser,
		Password: dbpassword,
		Database: "postgres",
	})
	defer func() error {
		if err := db.Close(); err != nil {
			return fmt.Errorf("error closing database %w", err)
		}
		return nil
	}()
	if err != nil {
		return fmt.Errorf("error opening database %w", err)
	}

	db.MustExec("drop database if exists " + destDB)
	db.MustExec("select pg_terminate_backend (pid) from pg_stat_activity where datname=$1", sourceDB)
	db.MustExec(fmt.Sprintf("create database %s with template %s owner %s", destDB, sourceDB, dbuser))

	return nil
}

// dbReset Database Reset DBUUID and remove MCode
func dbReset(dbhost, dbname, dbuser, dbpassword string) error {
	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Username: dbuser,
		Password: dbpassword,
		Database: dbname,
	})
	defer func() error {
		if err := db.Close(); err != nil {
			return fmt.Errorf("error closing database %w", err)
		}
		return nil
	}()
	if err != nil {
		return fmt.Errorf("error opening database %w", err)
	}

	db.MustExec("delete from ir_config_parameter where key='database.enterprise_code'")
	db.MustExec(`update ir_config_parameter
		set value=(select gen_random_uuid())
		where key = 'database.uuid'`,
	)
	db.MustExec(`insert into ir_config_parameter
	    (key,value,create_uid,create_date,write_uid,write_date) values
	    ('database.expiration_date',(current_date+'3 months'::interval)::timestamp,1,
	    current_timestamp,1,current_timestamp)
	    on conflict (key)
	    do update set value = (current_date+'3 months'::interval)::timestamp;`,
	)

	return nil
}

// restoreAddonsTar Restore Odoo DB addons folders
func restoreAddonsTar(addonsFile string) error {
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

func restoreAddonsTarNode(addonsFile string) error {
	odir := "/opt/odoo"
	source := filepath.Join(odir, "backups", addonsFile)
	dest := filepath.Join(odir, "addons")
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

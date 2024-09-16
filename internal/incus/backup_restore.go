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

	uid, err := IncusGetUid(project, "odoo")
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("could not get odoo uid %w", err)
	}

	if err := exec.Command("incus", "exec", project, "--user", uid, "-t",
		"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
		"oda", "backup",
	).Run(); err != nil {
		return fmt.Errorf("backup failed %w", err)
	}

	return nil
}

// AdminRestore Restore from backup file
func AdminRestore(any, move bool) error {
	if !IsProject() {
		return fmt.Errorf("not in a project directory")
	}
	_, project := GetProject()

	var backups []string
	var addons []string

	if any {
		backups, addons = GetOdooBackups("")
	} else {
		backups, addons = GetOdooBackups(project)
	}

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

	huh.NewSelect[string]().
		Title("Odoo Backup File").
		Options(backupOptions...).
		Value(&backupFile).
		Run()

	huh.NewSelect[string]().
		Title("Odoo Addon File").
		Options(addonOptions...).
		Value(&addonFile).
		Run()

	huh.NewConfirm().
		Title("Restore Project?").
		Value(&confirm).
		Run()

	if !confirm {
		fmt.Println("restore cancelled")
		return nil
	}

	if addonFile != "none" {
		fmt.Println("restore from addon file " + addonFile)
		if err := restoreAddonsTar(addonFile); err != nil {
			return fmt.Errorf("restore addons tar failed %w", err)
		}
	}

	fmt.Println("restore from backup file " + backupFile)
	if err := restoreDBTar(backupFile, move); err != nil {
		return fmt.Errorf("restore db tar failed %w", err)
	}

	return nil
}

// restoreDBTar Restore Odoo DB from backup
func restoreDBTar(backupFile string, moveDB bool) error {
	conf := GetConf()
	cwd, project := GetProject()
	dirs := GetDirs()
	source := filepath.Join(dirs.Project, "backups", backupFile)

	dbname := GetOdooConf(cwd, "db_name")
	dbhost := GetOdooConf(cwd, "db_host")
	dbuser := GetOdooConf(cwd, "db_user")
	dbpassword := GetOdooConf(cwd, "db_password")
	dbtemplate := GetOdooConf(cwd, "db_template")

	dbserver := dbhost
	if dbhost == "localhost" {
		dbserver = project
	}
	uid, err := IncusGetUid(dbserver, "postgres")
	if err != nil {
		return fmt.Errorf("could not get postgres uid %w", err)
	}

	// drop target database
	// fmt.Println("drop target database")
	if err := exec.Command("incus", "exec", dbserver, "--user", uid, "-t", "--",
		"dropdb", "--if-exists", "-U", "postgres", "-f", dbname,
	).Run(); err != nil {
		return fmt.Errorf("could not drop postgresql database %s error: %w", dbname, err)
	}
	// fmt.Println("dropped database " + dbname)

	// create new postgresql database
	// fmt.Println("create new postgresql database")
	if err := exec.Command("incus", "exec", dbserver, "--user", uid, "-t", "--",
		"createdb", "-U", "postgres",
		"--encoding", "unicode",
		"--lc-collate", "C",
		"-T", dbtemplate,
		"-O", dbuser, dbname,
	).Run(); err != nil {
		return fmt.Errorf("could not create postgresql database %s error: %w", dbname, err)
	}
	// fmt.Println("created database " + dbname)

	// restore postgresql database
	// fmt.Println("restore postgresql database")
	tarpgCmd := exec.Command("tar", "Oaxf", source, "./dump.sql")
	dbhostTarget := dbhost + "." + conf.Domain
	if dbhost == "localhost" {
		dbInstance, _ := GetInstance(project)
		fmt.Println(dbInstance)
		dbhostTarget = dbInstance.IP4
	}
	pgCmd := exec.Command("psql", "-h", dbhostTarget, "-U", dbuser, "--dbname", dbname, "-q")
	pgCmd.Env = append(pgCmd.Env, "PGPASSWORD="+dbpassword)

	// fmt.Println("tar", "Oaxf", source, "./dump.sql")
	// fmt.Println("PGPASSWORD="+dbpassword, "psql", "-h", dbhostTarget, "-U", dbuser, "--dbname", dbname, "-q")

	r, w := io.Pipe()
	tarpgCmd.Stdout = w
	pgCmd.Stdin = r

	tarpgCmd.Start()
	pgCmd.Start()
	tarpgCmd.Wait()
	w.Close()
	pgCmd.Wait()
	// fmt.Println("restored database " + dbname)

	// restore data filestore
	// fmt.Println("restore postgresql database")
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
	// fmt.Println("restored filestore " + dbname)

	// if not moveDB then reset DBUUID and remove MCode
	if !moveDB {
		fmt.Println("neutralize the database")
		if err := dbReset(dbhostTarget, dbname, dbuser, dbpassword); err != nil {
			return fmt.Errorf("db reset failed %w", err)
		}
	}
	// fmt.Println("reset database " + dbname)
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

	db.Exec("drop database if exists " + destDB)
	db.Exec("select pg_terminate_backend (pid) from pg_stat_activity where datname=$1", sourceDB)
	db.Exec(fmt.Sprintf("create database %s with template %s owner %s", destDB, sourceDB, dbuser))

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

	db.Exec("delete from ir_config_parameter where key='database.enterprise_code'")
	db.Exec(`update ir_config_parameter
		set value=(select gen_random_uuid())
		where key = 'database.uuid'`,
	)
	db.Exec(`insert into ir_config_parameter
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
		return fmt.Errorf("remove contents failed: %w", err)
	}
	cmd := exec.Command("tar",
		"axf", source, "-C", dest, ".",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract addon files failed: %w", err)
	}
	return nil
}

package oda

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Database struct contains sql pointer
type Database struct {
	Hostname string `json:"hostname,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	URI      string `json:"uri,omitempty"`
	*sqlx.DB
}

// OpenDatabase open database
func OpenDatabase(db Database) (*Database, error) {
	var err error
	db.GetURI()
	db.DB, err = sqlx.Open("pgx", db.URI)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}
	return &db, err
}

// GenURI generate db uri string
func (db *Database) GetURI() {
	port := 5432
	if db.Port != 0 {
		port = db.Port
	}
	db.URI = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		db.Username, db.Password, db.Hostname, port, db.Database)
}

func PgdbPgsql() error {
	conf := GetConf()
	uid, err := PgdbGetuid(conf.DBH)
	if err != nil {
		fmt.Println(err)
		return err
	}
	pgCmd := exec.Command("incus", "exec", conf.DBH, "--user", uid, "-t", "--", "psql")
	pgCmd.Stdin = os.Stdin
	pgCmd.Stdout = os.Stdout
	pgCmd.Stderr = os.Stderr
	if err := pgCmd.Run(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func PgdbStart() error {
	conf := GetConf()
	container, err := GetContainer(conf.DBH)
	if err != nil {
		return err
	}
	if container.State != "RUNNING" {
		if err := IncusStart(conf.DBH); err != nil {
			return err
		}
	}
	fmt.Println(conf.DBH, "started")
	return nil
}

func PgdbStop() error {
	conf := GetConf()
	container, err := GetContainer(conf.DBH)
	if err != nil {
		return err
	}
	if container.State != "STOPPED" {
		if err := IncusStop(conf.DBH); err != nil {
			return err
		}
	}
	fmt.Println(conf.DBH, "stopped")
	return nil
}

func PgdbRestart() error {
	if err := PgdbStop(); err != nil {
		return err
	}
	if err := PgdbStart(); err != nil {
		return err
	}
	return nil
}

func PgdbFullReset() error {
	conf := GetConf()
	confim := AreYouSure("reset the " + conf.DBH + " database server")
	if !confim {
		return fmt.Errorf("reset of the database server")
	}

	if err := PgdbStop(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database server " + conf.DBH + " stopped")

	if err := IncusDelete(conf.DBH); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database server " + conf.DBH + " data cleared")

	if err := IncusLaunch(conf.DBH, conf.DBImage); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database server " + conf.DBH + " image launched")

	if err := exec.Command("incus", "exec", conf.DBH, "-t", "--", "apt", "upgrade", "-y").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	if err := exec.Command("incus", "exec", conf.DBH, "-t", "--", "apt", "install", "-y", "postgresql", "postgresql-contrib", "pgbouncer").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Database server " + conf.DBH + " packages installed")

	uid, err := PgdbGetuid(conf.DBH)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err := exec.Command("incus", "exec", conf.DBH, "--user", uid, "-t", "--", "psql", "-c", "CREATE ROLE "+conf.DBUsername+" WITH CREATEDB NOSUPERUSER ENCRYPTED PASSWORD '"+conf.DBUserpass+"' LOGIN;").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Database server " + conf.DBUsername + " role created")

	// /etc/pgbouncer/pgbouncer.ini
	CONFIG, _ := os.UserConfigDir()
	PGBOUNCERINI := filepath.Join(CONFIG, "oda", "pgbouncer.ini")
	if err := exec.Command("incus", "file", "push", PGBOUNCERINI, conf.DBH+"/etc/pgbouncer/pgbouncer.ini").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	// /etc/pgbouncer/userlist.txt
	if err := exec.Command("incus", "exec", conf.DBH, "--user", uid, "-t", "--", "psql", "-c", "COPY(SELECT '\"'||rolname||'\" \"'||coalesce(rolpassword,'')||'\"' from pg_authid) TO '/etc/pgbouncer/userlist.txt';").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Database server " + conf.DBUsername + " pgbouncer userlist.txt created")

	if err := PgdbRestart(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database server " + conf.DBH + " restarted")

	return nil
}

func PgdbGetuid(dbhost string) (string, error) {
	out, err := exec.Command("incus", "exec", dbhost, "-t", "--", "grep", "^postgres", "/etc/passwd").Output()
	if err != nil {
		return "", err
	}
	return strings.Split(string(out), ":")[2], nil
}

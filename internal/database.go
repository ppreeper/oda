package oda

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
	fmt.Println("Database opened")
	if err = db.Ping(); err != nil {
		fmt.Println("Database ping failed")
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
	uid, err := IncusGetUid(conf.DBHost, "postgres")
	if err != nil {
		fmt.Println(err)
		return err
	}
	pgCmd := exec.Command("incus", "exec", conf.DBHost, "--user", uid, "-t", "--", "psql")
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
	container, err := GetContainer(conf.DBHost)
	if err != nil {
		return err
	}
	if container.State != "RUNNING" {
		if err := IncusStart(conf.DBHost); err != nil {
			return err
		}
	}
	fmt.Println(conf.DBHost, "started")
	return nil
}

func PgdbStop() error {
	conf := GetConf()
	container, err := GetContainer(conf.DBHost)
	if err != nil {
		return err
	}
	if container.State != "STOPPED" {
		if err := IncusStop(conf.DBHost); err != nil {
			return err
		}
	}
	fmt.Println(conf.DBHost, "stopped")
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
	confim := AreYouSure("reset the " + conf.DBHost + " database server")
	if !confim {
		return fmt.Errorf("reset of the database server")
	}

	if err := PgdbStop(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database server " + conf.DBHost + " stopped")

	if err := IncusDelete(conf.DBHost); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database server " + conf.DBHost + " data cleared")

	if err := IncusLaunch(conf.DBHost, conf.DBImage); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database server " + conf.DBHost + " image launched")

	if err := exec.Command("incus", "exec", conf.DBHost, "-t", "--", "apt", "upgrade", "-y").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	if err := exec.Command("incus", "exec", conf.DBHost, "-t", "--", "apt", "install", "-y", "postgresql", "postgresql-contrib", "pgbouncer").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Database server " + conf.DBHost + " packages installed")

	uid, err := IncusGetUid(conf.DBHost, "postgres")
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err := exec.Command("incus", "exec", conf.DBHost, "--user", uid, "-t", "--", "psql", "-c", "ALTER ROLE postgres WITH ENCRYPTED PASSWORD '"+conf.DBPass+"';").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Database server postgres role altered")

	if err := exec.Command("incus", "exec", conf.DBHost, "--user", uid, "-t", "--", "psql", "-c", "CREATE ROLE "+conf.DBUsername+" WITH CREATEDB NOSUPERUSER ENCRYPTED PASSWORD '"+conf.DBUserpass+"' LOGIN;").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Database server " + conf.DBUsername + " role created")

	CONFIG, _ := os.UserConfigDir()
	// /etc/postgresql/15/main/conf.d/postgresql.conf
	POSTGRESQLCONF := filepath.Join(CONFIG, "oda", "postgresql.conf")
	if err := exec.Command("incus", "file", "push", POSTGRESQLCONF, conf.DBHost+"/etc/postgresql/15/main/conf.d/postgresql.conf").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	// /etc/postgresql/15/main/pg_hba.conf
	PGHBACONF := filepath.Join(CONFIG, "oda", "pg_hba.conf")
	if err := exec.Command("incus", "file", "push", PGHBACONF, conf.DBHost+"/etc/postgresql/15/main/pg_hba.conf").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	// /etc/pgbouncer/pgbouncer.ini
	PGBOUNCERINI := filepath.Join(CONFIG, "oda", "pgbouncer.ini")
	if err := exec.Command("incus", "file", "push", PGBOUNCERINI, conf.DBHost+"/etc/pgbouncer/pgbouncer.ini").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	// /etc/pgbouncer/userlist.txt
	if err := exec.Command("incus", "exec", conf.DBHost, "--user", uid, "-t", "--", "psql", "-c", "COPY(SELECT '\"'||rolname||'\" \"'||coalesce(rolpassword,'')||'\"' from pg_authid) TO '/etc/pgbouncer/userlist.txt';").Run(); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Database server " + conf.DBUsername + " pgbouncer userlist.txt created")

	if err := PgdbRestart(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database server " + conf.DBHost + " restarted")

	return nil
}

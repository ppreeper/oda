package server

import (
	"fmt"
	"io"
	"os/exec"

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

type OdooDB struct {
	Hostname string `json:"hostname,omitempty"`
	Port     string `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Template string `json:"template,omitempty"`
}

func (o *OdooDB) DropDatabase() error {
	dropdbCmd := exec.Command("dropdb", "--if-exists",
		"-h", o.Hostname,
		"-p", o.Port,
		"-U", o.Username,
		"-f", o.Database,
	)
	dropdbCmd.Env = append(dropdbCmd.Env, "PGPASSWORD="+o.Password)
	if err := dropdbCmd.Run(); err != nil {
		return fmt.Errorf("could not drop postgresql database %s error: %w", o.Database, err)
	}
	return nil
}

func (o *OdooDB) CreateDatabase() error {
	createdbCmd := exec.Command("createdb",
		"-h", o.Hostname,
		"-p", o.Port,
		"-U", o.Username,
		"--encoding", "unicode",
		"--lc-collate", "C",
		"-T", o.Template,
		"-O", o.Username,
		o.Database,
	)
	createdbCmd.Env = append(createdbCmd.Env, "PGPASSWORD="+o.Password)
	if err := createdbCmd.Run(); err != nil {
		return fmt.Errorf("could not create postgresql database %s error: %w", o.Database, err)
	}
	return nil
}

func (o *OdooDB) RestoreDatabase(source string) error {
	tarpgCmd := exec.Command("tar", "Oaxf", source, "./dump.sql")
	pgCmd := exec.Command("psql", "-h", o.Hostname, "-U", o.Username, "--dbname", o.Database, "-q")
	pgCmd.Env = append(pgCmd.Env, "PGPASSWORD="+o.Password)

	r, w := io.Pipe()
	tarpgCmd.Stdout = w
	pgCmd.Stdin = r

	tarpgCmd.Start()
	pgCmd.Start()
	tarpgCmd.Wait()
	w.Close()
	return pgCmd.Wait()
}

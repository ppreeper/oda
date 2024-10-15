package internal

import (
	"fmt"
	"os"

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
		fmt.Fprintln(os.Stderr, "Database ping failed")
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

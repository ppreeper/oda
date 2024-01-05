package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
		if err != nil {
			return nil, fmt.Errorf("cannot open database: %w", err)
		}
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

func pgdbPgsql() error {
	conf := GetConf()
	podCmd := exec.Command("podman",
		"exec", "-it", conf.DBHost,
		"psql", "-U", "postgres",
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		return err
	}
	return nil
}

func pgdbStart() error {
	conf := GetConf()

	pods, _ := getPods(true)
	for _, pod := range pods {
		if strings.Contains(pod.Name, conf.DBHost) &&
			strings.HasPrefix(pod.Status, "Up") {
			fmt.Println("Database already running")
			return nil
		}
		if strings.Contains(pod.Name, conf.DBHost) &&
			(strings.HasPrefix(pod.Status, "Created") || strings.HasPrefix(pod.Status, "Exited")) {
			instanceStop()
		}
	}

	if err := exec.Command("podman",
		"run", "--name", conf.DBHost,
		"-p", conf.DBPort+":5432",
		"-e", "POSTGRES_PASSWORD="+conf.DBPass,
		"-v", conf.DBHost+":/var/lib/postgresql/data",
		"--rm", "-d", "docker.io/postgres:15-alpine",
	).Run(); err != nil {
		return err
	}
	return nil
}

func pgdbStop() error {
	conf := GetConf()
	if err := exec.Command("podman", "stop", conf.DBHost).Run(); err != nil {
		fmt.Println("stopping: ", err)
	}
	if err := exec.Command("podman", "rm", conf.DBHost).Run(); err != nil {
		fmt.Println("removing: ", err)
	}
	fmt.Println(conf.DBHost, "stopped")
	return nil
}

func pgdbRestart() error {
	if err := pgdbStop(); err != nil {
		return err
	}
	if err := pgdbStart(); err != nil {
		return err
	}
	return nil
}

func pgdbFullReset() error {
	conf := GetConf()
	if err := pgdbStop(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database stopped")
	if err := exec.Command("podman",
		"volume", "rm ", conf.DBHost,
	).Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Volume removed")
	if err := pgdbStart(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Database started")

	// Delay for db warmup
	time.Sleep(2 * time.Second)

	started := false
	for !started {
		pods, _ := exec.Command("podman", "ps", "--format", "{{.Names}}").Output()
		podStrings := strings.Split(string(pods), "\n")
		for _, pod := range podStrings {
			if pod == conf.DBHost {
				started = true
			}
		}
		time.Sleep(1 * time.Second)
	}

	db, err := OpenDatabase(Database{
		Hostname: conf.DBHost,
		Username: "postgres",
		Password: conf.DBPass,
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

	db.MustExec(fmt.Sprintf("create role %s with createdb login password %s ;", conf.DBUsername, conf.DBUserpass))

	return nil
}

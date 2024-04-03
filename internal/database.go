package oda

import (
	"fmt"
	"os"
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

	// PostgreSQL Config
	POSTGRESCONF()

	// PGHBA Config
	PGHBACONF()

	// PGBouncer Config
	PGBouncer()

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

func POSTGRESCONF() error {
	conf := GetConf()
	localFile := "/tmp/" + conf.DBHost + "-postgresql.conf"

	f, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString("listen_addresses = '*'" + "\n")
	f.WriteString("port = 5432" + "\n")
	f.WriteString("max_connections = 100" + "\n")
	f.WriteString("shared_buffers = 128MB" + "\n")
	f.WriteString("effective_cache_size = 4GB" + "\n")
	f.WriteString("maintenance_work_mem = 512MB" + "\n")
	f.WriteString("checkpoint_completion_target = 0.9" + "\n")
	f.WriteString("wal_buffers = 16MB" + "\n")
	f.WriteString("default_statistics_target = 100" + "\n")
	f.WriteString("random_page_cost = 1.1" + "\n")
	f.WriteString("effective_io_concurrency = 200" + "\n")
	f.WriteString("work_mem = 4MB" + "\n")
	f.WriteString("min_wal_size = 1GB" + "\n")
	f.WriteString("max_wal_size = 2GB" + "\n")
	f.WriteString("max_worker_processes = 8" + "\n")
	f.WriteString("max_parallel_workers_per_gather = 4" + "\n")
	f.WriteString("max_parallel_workers = 8" + "\n")
	f.WriteString("max_parallel_maintenance_workers = 4" + "\n")

	if err := exec.Command("incus", "file", "push", localFile, conf.DBHost+"/etc/postgresql/15/main/conf.d/postgresql.conf").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	os.Remove(localFile)

	return nil
}

func PGHBACONF() error {
	conf := GetConf()
	localFile := "/tmp/" + conf.DBHost + "-pg_hba.conf"

	f, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString("local  all          postgres                      peer" + "\n")
	f.WriteString("local  all          all                           peer" + "\n")
	f.WriteString("host   all          all       0.0.0.0/0           scram-sha-256" + "\n")
	f.WriteString("host   all          all       127.0.0.1/32        scram-sha-256" + "\n")
	f.WriteString("host   all          all       ::1/128             scram-sha-256" + "\n")
	f.WriteString("local  replication  all                           peer" + "\n")
	f.WriteString("host   replication  all       127.0.0.1/32        scram-sha-256" + "\n")
	f.WriteString("host   replication  all       ::1/128             scram-sha-256" + "\n")

	if err := exec.Command("incus", "file", "push", localFile, conf.DBHost+"/etc/postgresql/15/main/pg_hba.conf").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	os.Remove(localFile)

	return nil
}

func PGBouncer() error {
	conf := GetConf()
	localFile := "/tmp/" + conf.DBHost + "-pgbouncer.ini"

	f, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString("[databases]" + "\n")
	f.WriteString("* = host=127.0.0.1 port=5432" + "\n")
	f.WriteString("[pgbouncer]" + "\n")
	f.WriteString("logfile = /var/log/postgresql/pgbouncer.log" + "\n")
	f.WriteString("pidfile = /var/run/postgresql/pgbouncer.log" + "\n")
	f.WriteString("unix_socket_dir = /var/run/postgresql" + "\n")
	f.WriteString("listen_addr = *" + "\n")
	f.WriteString("listen_port = 6432" + "\n")
	f.WriteString("auth_type = md5" + "\n")
	f.WriteString("auth_file = /etc/pgbouncer/userlist.txt" + "\n")
	f.WriteString("max_client_conn=1000" + "\n")
	f.WriteString("default_pool_size=25" + "\n")
	f.WriteString("reserve_pool_size=5" + "\n")
	f.WriteString("pool_mode = transaction" + "\n")
	f.WriteString("server_reset_query = DISCARD ALL" + "\n")
	f.WriteString("ignore_startup_parameters = extra_float_digits" + "\n")

	if err := exec.Command("incus", "file", "push", localFile, conf.DBHost+"/etc/pgbouncer/pgbouncer.ini").Run(); err != nil {
		fmt.Println(err)
		return err
	}

	os.Remove(localFile)

	return nil
}

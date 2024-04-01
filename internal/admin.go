package oda

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func AdminInit() error {
	HOME, _ := os.UserHomeDir()

	conf := *NewConf()

	// Repo
	REPO := filepath.Join(HOME, "workspace/repos/odoo")
	if _, err := os.Stat(REPO); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(REPO), 0o755)
	}
	conf.Repo = REPO

	// Project
	PROJECT := filepath.Join(HOME, "workspace/odoo")
	if _, err := os.Stat(filepath.Join(PROJECT, "backups")); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(PROJECT, "backups"), 0o755)
	}
	conf.Project = PROJECT

	confRec, _ := json.Marshal(conf)
	var confMap map[string]any
	json.Unmarshal(confRec, &confMap)

	keys := []string{}
	for field := range confMap {
		keys = append(keys, field)
	}
	sort.Strings(keys)

	// ODA Config
	ODAConfig(keys, confMap)

	// PostgreSQL Config
	POSTGRESCONF()

	// PGHBA Config
	PGHBACONF()

	// PGBouncer Config
	PGBouncer()

	return nil
}

func ODAConfig(keys []string, confMap map[string]any) error {
	CONFIG, _ := os.UserConfigDir()
	// ODA Config
	ODOOCONF := filepath.Join(CONFIG, "oda", "oda.conf")
	if _, err := os.Stat(ODOOCONF); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(CONFIG, "oda"), 0o755)
		f, err := os.Create(ODOOCONF)
		if err != nil {
			return err
		}
		defer f.Close()
		for _, field := range keys {
			f.WriteString(fmt.Sprintf("%s=%s\n", field, confMap[field]))
		}
	}
	return nil
}

func POSTGRESCONF() error {
	CONFIG, _ := os.UserConfigDir()
	POSTGRESCONF := filepath.Join(CONFIG, "oda", "postgresql.conf")
	if _, err := os.Stat(POSTGRESCONF); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(CONFIG, "oda"), 0o755)
		f, err := os.Create(POSTGRESCONF)
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
	}

	return nil
}

func PGHBACONF() error {
	CONFIG, _ := os.UserConfigDir()
	PGHBACONF := filepath.Join(CONFIG, "oda", "pg_hba.conf")
	if _, err := os.Stat(PGHBACONF); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(CONFIG, "oda"), 0o755)
		f, err := os.Create(PGHBACONF)
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
	}
	return nil
}

func PGBouncer() error {
	CONFIG, _ := os.UserConfigDir()
	// PGBouncer Config
	PGBOUNCERINI := filepath.Join(CONFIG, "oda", "pgbouncer.ini")
	if _, err := os.Stat(PGBOUNCERINI); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(CONFIG, "oda"), 0o755)
		f, err := os.Create(PGBOUNCERINI)
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
	}
	return nil
}

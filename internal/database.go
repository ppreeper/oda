package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/oda/ui"
	"github.com/ppreeper/odoorpc/odoojrpc"
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

func (o *ODA) DBFullReset() error {
	odaConf, _ := config.LoadOdaConfig()
	dbHost := odaConf.Database.Host

	confim := ui.AreYouSure("reset the " + dbHost + " database server")
	if !confim {
		fmt.Fprintln(os.Stderr, "reset of the database server")
		return nil
	}

	if err := o.DBCreateScript(fmt.Sprintf("%d", odaConf.Database.Version)); err != nil {
		fmt.Fprintln(os.Stderr, "error creating database script 1", err)
		return nil
	}

	return nil
}

func (o *ODA) DBLogs() error {
	odaConf, _ := config.LoadOdaConfig()
	dbHost := odaConf.Database.Host

	podCmd := exec.Command("incus",
		"exec", dbHost, "-t", "--",
		"journalctl", "-f",
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		return nil
	}
	return nil
}

func (o *ODA) DBEXEC() error {
	odaConf, _ := config.LoadOdaConfig()

	dbHost := odaConf.Database.Host

	incusCmd := exec.Command("incus", "exec", dbHost, "-t", "--", "/bin/bash")
	incusCmd.Stdin = os.Stdin
	incusCmd.Stdout = os.Stdout
	incusCmd.Stderr = os.Stderr
	if err := incusCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("error executing %v"), err)
		return nil
	}
	return nil
}

func (o *ODA) DBPSQL() error {
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	dbHost := odaConf.Database.Host
	dbuser := "postgres"
	dbpassword := odaConf.Database.Password
	dbName := "postgres"

	uid, err := inc.IncusGetUid(dbHost, dbuser)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("could not get postgres uid %v"), err)
		return nil
	}

	incusCmd := exec.Command("incus", "exec", dbHost, "--user", uid,
		"--env", "PGPASSWORD="+dbpassword, "-t", "--",
		"psql", "-h", "127.0.0.1", "-U", dbuser, dbName,
	)
	incusCmd.Stdin = os.Stdin
	incusCmd.Stdout = os.Stdout
	incusCmd.Stderr = os.Stderr
	if err := incusCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error instance psql %w", err)
	}
	return nil
}

func (o *ODA) DBStart() error {
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)
	dbHost := odaConf.Database.Host
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Starting", dbHost))
	inc.SetInstanceState(dbHost, "start")
	instanceStatus := inc.GetInstanceState(dbHost)
	fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(dbHost, instanceStatus.Metadata.Status))
	return nil
}

func (o *ODA) DBStop() error {
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)
	dbHost := odaConf.Database.Host
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Stopping", dbHost))
	inc.SetInstanceState(dbHost, "stop")
	instanceStatus := inc.GetInstanceState(dbHost)
	fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(dbHost, instanceStatus.Metadata.Status))
	return nil
}

func (o *ODA) DBRestart() error {
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)
	dbHost := odaConf.Database.Host
	// Stop
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Stopping", dbHost))
	inc.SetInstanceState(dbHost, "stop")
	instanceStatus := inc.GetInstanceState(dbHost)
	fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(dbHost, instanceStatus.Metadata.Status))
	// Start
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Starting", dbHost))
	inc.SetInstanceState(dbHost, "start")
	instanceStatus = inc.GetInstanceState(dbHost)
	fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(dbHost, instanceStatus.Metadata.Status))
	return nil
}

func (o *ODA) Query() error {
	if !IsProject() {
		return nil
	}

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)
	cwd, project := lib.GetProject()
	odooConf, _ := config.LoadOdooConfig(cwd)
	instance, err := inc.GetInstance(project)
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("error getting instance %v\n"), err)
		return nil
	}

	dbname := odooConf.DbName

	oc := odoojrpc.NewOdoo().
		WithHostname(instance.IP4).
		WithPort(8069).
		WithDatabase(dbname).
		WithUsername(o.Q.Username).
		WithPassword(o.Q.Password).
		WithSchema("http")

	err = oc.Login()
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("error logging in %v\n"), err)
		return nil
	}

	umdl := strings.Replace(o.Q.Model, "_", ".", -1)

	fields := parseFields(o.Q.Fields)
	if o.Q.Count {
		fields = []string{"id"}
	}

	filtp, err := parseFilter(o.Q.Filter)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("filter parse error"))
	}

	rr, err := oc.SearchRead(umdl, o.Q.Offset, o.Q.Limit, fields, filtp)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("searchread error"))
		return nil
	}
	if o.Q.Count {
		fmt.Fprintf(os.Stderr, ui.SubStepStyle.Render("records: %d\n"), len(rr))
	} else {
		jsonStr, err := json.MarshalIndent(rr, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("output format error\n"))
			return nil
		}
		fmt.Fprintln(os.Stderr, string(jsonStr))
	}
	return nil
}

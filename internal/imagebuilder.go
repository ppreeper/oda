package internal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/oda/ui"
)

// base odoo image
func (o *ODA) BaseCreateScript(version string) error {
	fmt.Println("BaseCreateScript")
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Creating base image for Odoo version", version))
	branchConfig := config.GetVersion(version)
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	// TODO: add cpu and memory limits
	inc.CreateInstance(branchConfig.InstanceName, branchConfig.Image, odaConf.Incus.LimitCPU, odaConf.Incus.LimitMemory)

	inc.WaitForInstance(branchConfig.InstanceName, "RUNNING")

	roleUpdateScript(branchConfig.InstanceName)

	rolePreeperRepo(branchConfig.InstanceName, o.EmbedFS)

	inc.IncusExecVerbose(branchConfig.InstanceName, "odas", "welcome")

	// roleGupScript(config.InstanceName)

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add common system packages to", branchConfig.InstanceName))
	if err := aptInstall(branchConfig.InstanceName, branchConfig.BaselinePackages...); err != nil {
		return fmt.Errorf("apt-get install baseline failed %w", err)
	}

	rolePostgresqlRepo(branchConfig.InstanceName)

	rolePostgresqlClient(branchConfig.InstanceName, fmt.Sprintf("%d", odaConf.Database.Version))

	roleWkhtmltopdf(branchConfig.InstanceName)

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("add odoo dependencies to", branchConfig.InstanceName))
	if err := aptInstall(branchConfig.InstanceName, branchConfig.Odoobase...); err != nil {
		return fmt.Errorf("apt-get install baseline failed %w", err)
	}

	npmInstall(branchConfig.InstanceName, "rtlcss")

	roleGeoIP2DB(branchConfig.InstanceName)

	rolePaperSize(branchConfig.InstanceName)

	roleOdooUser(branchConfig.InstanceName)

	roleOdooDirs(branchConfig.InstanceName, branchConfig.Repos)

	roleCaddy(branchConfig.InstanceName)

	roleCaddyService(branchConfig.InstanceName, o.EmbedFS)

	roleOdooService(branchConfig.InstanceName, o.EmbedFS)

	inc.SetInstanceState(branchConfig.InstanceName, "stop")

	return nil
}

func (o *ODA) DBCreateScript(version string) error {
	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)
	dbHost := odaConf.Database.Host
	dbUsername := odaConf.Database.Username
	dbPassword := odaConf.Database.Password

	// Destroy Database Instance
	inc.SetInstanceState(dbHost, "stop")
	inc.DeleteInstance(dbHost)
	time.Sleep(5 * time.Second) // debounce timer

	// Create Database Instance
	inc.CreateInstance(dbHost, odaConf.Database.Image, 4, "4GiB")
	inc.SetInstanceState(dbHost, "start")
	time.Sleep(5 * time.Second) // debounce timer

	// Start Installation Process
	roleUpdateScript(dbHost)

	// PostgreSQL Config
	rolePostgresqlRepo(dbHost)
	rolePostgresqlServer(dbHost, fmt.Sprintf("%d", odaConf.Database.Version))

	// postgresql.conf
	rolePostgresqlConf(dbHost, o.EmbedFS)

	// pg_hba.conf
	rolePghbaConf(dbHost, o.EmbedFS)

	// Setup User Roles
	uid, err := inc.IncusGetUid(dbHost, "postgres")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get postgres uid %w", err)
		return nil
	}

	// Alter postgres Role
	err = exec.Command("incus", "exec", dbHost, "--user", uid, "-t", "--",
		"psql", "-c", "ALTER ROLE postgres WITH ENCRYPTED PASSWORD '"+dbPassword+"';").Run()
	if err != nil {
		if errors.Is(err, exec.ErrWaitDelay) {
			fmt.Fprintln(os.Stderr, "failed to alter postgres role %w", err)
			return nil
		}
	}

	// add pg_stat_statements to postgres database
	err = exec.Command("incus", "exec", dbHost, "--user", uid, "-t", "--",
		"psql", "-c", "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;").Run()
	if err != nil {
		if errors.Is(err, exec.ErrWaitDelay) {
			fmt.Fprintln(os.Stderr, "failed to create extension pg_stat_statements %w", err)
			return nil
		}
	}

	// Create odoo Role
	err = exec.Command("incus", "exec", dbHost, "--user", uid, "-t", "--",
		"psql", "-c", "CREATE ROLE "+dbUsername+" WITH CREATEDB NOSUPERUSER ENCRYPTED PASSWORD '"+dbPassword+"' LOGIN;").Run()
	if err != nil {
		if errors.Is(err, exec.ErrWaitDelay) {
			fmt.Fprintln(os.Stderr, "failed to create role %w", err)
			return nil
		}
	}

	// Restart PostgreSQL
	inc.SetInstanceState(dbHost, "stop")
	time.Sleep(5 * time.Second) // debounce timer
	inc.SetInstanceState(dbHost, "start")

	return nil
}

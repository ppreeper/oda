/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dbFullresetCmd = &cobra.Command{
	Use:   "fullreset",
	Short: "database fullreset",
	Long:  `database fullreset`,
	Run: func(cmd *cobra.Command, args []string) {
		dbHost := viper.GetString("database.host")

		confim := AreYouSure("reset the " + dbHost + " database server")
		if !confim {
			fmt.Fprintln(os.Stderr, "reset of the database server")
			return
		}

		SetInstanceState(dbHost, "stop")
		fmt.Fprintln(os.Stderr, "postgresql server "+dbHost+" stopped")

		DeleteInstance(dbHost)
		fmt.Fprintln(os.Stderr, "postgresql server "+dbHost+" data cleared")

		time.Sleep(5 * time.Second)

		CreateInstance(dbHost, OdooDatabase.Image)
		fmt.Fprintln(os.Stderr, "postgresql server "+dbHost+" image launched")
		WaitForInstance(dbHost, "start")

		roleUpdateScript(dbHost)
		rolePostgresqlRepo(dbHost)
		aptInstall(dbHost, "postgresql-"+OdooDatabase.Version, "pgbouncer")

		// Setup User Roles
		uid, err := IncusGetUid(dbHost, "postgres")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to get postgres uid %w", err)
			return
		}
		dbUsername := viper.GetString("database.username")
		dbPassword := viper.GetString("database.password")

		err = exec.Command("incus", "exec", dbHost, "--user", uid, "-t", "--",
			"psql", "-c", "ALTER ROLE postgres WITH ENCRYPTED PASSWORD '"+dbPassword+"';").Run()
		if err != nil {
			if errors.Is(err, exec.ErrWaitDelay) {
				fmt.Fprintln(os.Stderr, "failed to alter postgres role %w", err)
				return
			}
		}
		fmt.Fprintln(os.Stderr, "postgresql server postgres role altered")

		err = exec.Command("incus", "exec", dbHost, "--user", uid, "-t", "--",
			"psql", "-c", "CREATE ROLE "+dbUsername+" WITH CREATEDB NOSUPERUSER ENCRYPTED PASSWORD '"+dbPassword+"' LOGIN;").Run()
		if err != nil {
			if errors.Is(err, exec.ErrWaitDelay) {
				fmt.Fprintln(os.Stderr, "failed to create role %w", err)
				return
			}
		}
		fmt.Fprintln(os.Stderr, "postgresql server "+dbUsername+" role created")

		// PostgreSQL Config
		POSTGRESCONF()

		// PGHBA Config
		PGHBACONF()

		SetInstanceState(dbHost, "stop")
		SetInstanceState(dbHost, "start")
		fmt.Fprintln(os.Stderr, "postgresql server "+dbHost+" restarted")
	},
}

func init() {
	dbCmd.AddCommand(dbFullresetCmd)
}

func POSTGRESCONF() error {
	dbHost := viper.GetString("database.host")
	dbVersion := viper.GetString("database.version")

	localFile := "/tmp/" + dbHost + "-postgresql.conf"

	fo, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to create postgresql.conf: %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/postgresql.conf")
	cobra.CheckErr(err)
	err = t.Execute(fo, data)
	cobra.CheckErr(err)

	// debian/12
	if err := exec.Command("incus", "file", "push", localFile, dbHost+"/etc/postgresql/"+dbVersion+"/main/conf.d/postgresql.conf").Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("failed to push postgresql.conf: %w", err)
	}

	os.Remove(localFile)

	return nil
}

func PGHBACONF() error {
	dbHost := viper.GetString("database.host")
	dbVersion := viper.GetString("database.version")

	localFile := "/tmp/" + dbHost + "-pg_hba.conf"

	fo, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to create pg_hba.conf: %w", err)
	}
	defer fo.Close()

	data := map[string]string{}
	t, err := template.ParseFS(embedFS, "templates/pg_hba.conf")
	cobra.CheckErr(err)
	err = t.Execute(fo, data)
	cobra.CheckErr(err)

	// debian/12
	if err := exec.Command("incus", "file", "push", localFile, dbHost+"/etc/postgresql/"+dbVersion+"/main/pg_hba.conf").Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("failed to push pg_hba.conf: %w", err)
	}

	os.Remove(localFile)

	return nil
}

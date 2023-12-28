package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

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
		return err
	}
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
	if err := exec.Command("podman",
		"exec", "-it", "--user", "postgres", conf.DBHost,
		"psql", "-c",
		"create role "+conf.DBUsername+" with createdb login password '"+conf.DBUserpass+"';",
	).Run(); err != nil {
		return err
	}
	fmt.Println("User created")
	return nil
}

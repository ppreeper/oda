package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/ppreeper/passhash"
)

func AdminUsername() error {
	var user1, user2 string
	huh.NewInput().
		Title("Please enter  the new admin username:").
		Prompt(">").
		Value(&user1).
		Run()
	huh.NewInput().
		Title("Please verify the new admin username:").
		Prompt(">").
		Value(&user2).
		Run()

	if user1 == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if user1 != user2 {
		return fmt.Errorf("usernames entered do not match")
	}

	dbhost := GetOdooConf("", "db_host")
	dbname := GetOdooConf("", "db_name")
	dbuser := GetOdooConf("", "db_user")
	dbpassword := GetOdooConf("", "db_password")

	dbport, err := strconv.Atoi(GetOdooConf("", "db_port"))
	if err != nil {
		return fmt.Errorf("error getting port %w", err)
	}

	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Port:     dbport,
		Username: dbuser,
		Password: dbpassword,
		Database: dbname,
	})
	if err != nil {
		return fmt.Errorf("error opening database %w", err)
	}
	defer func() error {
		if err := db.Close(); err != nil {
			return fmt.Errorf("error closing database %w", err)
		}
		return nil
	}()

	// Write username to database
	_, err = db.Exec("update res_users set login=$1 where id=2;",
		strings.TrimSpace(string(user1)))
	if err != nil {
		return fmt.Errorf("error updating username %w", err)
	}

	fmt.Println("Admin username changed to", user1)
	return nil
}

func AdminPassword() error {
	var password1, password2 string
	huh.NewInput().
		Title("Please enter  the admin password:").
		Prompt(">").
		Password(true).
		Value(&password1).
		Run()
	huh.NewInput().
		Title("Please verify the admin password:").
		Prompt(">").
		Password(true).
		Value(&password2).
		Run()
	if password1 == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if password1 != password2 {
		return fmt.Errorf("passwords entered do not match")
	}
	var confirm bool
	huh.NewConfirm().
		Title("Are you sure you want to change the admin password?").
		Affirmative("yes").
		Negative("no").
		Value(&confirm).
		Run()
	if !confirm {
		return fmt.Errorf("password change cancelled")
	}

	dbhost := GetOdooConf("", "db_host")
	dbname := GetOdooConf("", "db_name")
	dbuser := GetOdooConf("", "db_user")
	dbpassword := GetOdooConf("", "db_password")

	dbport, err := strconv.Atoi(GetOdooConf("", "db_port"))
	if err != nil {
		return fmt.Errorf("error getting port %w", err)
	}

	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Port:     dbport,
		Username: dbuser,
		Password: dbpassword,
		Database: dbname,
	})
	if err != nil {
		return fmt.Errorf("error opening database %w", err)
	}
	defer func() error {
		if err := db.Close(); err != nil {
			return fmt.Errorf("error closing database %w", err)
		}
		return nil
	}()

	// Write password to database
	passkey, err := passhash.MakePassword(password1, 0, "")
	if err != nil {
		fmt.Println("password hashing error", err)
	}
	_, err = db.Exec("update res_users set password=$1 where id=2;",
		strings.TrimSpace(string(passkey)))
	if err != nil {
		return fmt.Errorf("error updating password %w", err)
	}

	fmt.Println("admin password changed")
	return nil
}

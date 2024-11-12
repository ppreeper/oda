package internal

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/passhash"
)

func (o *ODA) UpdateUser() error {
	if !IsProject() {
		return nil
	}

	odaConf, _ := config.LoadOdaConfig()
	// inc := incus.NewIncus(odaConf)
	cwd, _ := lib.GetProject()
	odooConf, _ := config.LoadOdooConfig(cwd)
	dbport, _ := strconv.Atoi(odooConf.DbPort)

	// setup db connection
	db, err := OpenDatabase(Database{
		Hostname: odooConf.DbHost + "." + odaConf.System.Domain,
		Port:     dbport,
		Username: odooConf.DbUser,
		Password: odooConf.DbPassword,
		Database: odooConf.DbName,
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

	getUsersQuery := "select id,company_id,partner_id,login from res_users where active=true order by login;"
	type User struct {
		ID        int    `db:"id"`
		CompanyID int    `db:"company_id"`
		PartnerID int    `db:"partner_id"`
		Login     string `db:"login"`
	}
	users := []User{}
	stmt, err := db.Preparex(getUsersQuery)
	if err != nil {
		fmt.Println("error preparing query", err)
	}
	err = stmt.Select(&users)
	if err != nil {
		fmt.Println("error getting users", err)
	}
	if len(users) == 0 {
		fmt.Println("No active users found")
	}

	usernames := []huh.Option[int]{}
	for _, user := range users {
		usernames = append(usernames, huh.NewOption(fmt.Sprintf("%d\t%s", user.ID, user.Login), user.ID))
	}

	var confirm bool
	var userid int
	formUser := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Odoo User").
				Options(usernames...).
				Value(&userid),
			huh.NewConfirm().
				Title("Select User").
				Value(&confirm),
		),
	)
	if err := formUser.Run(); err != nil {
		return fmt.Errorf("error updating user %w", err)
	}
	if !confirm {
		fmt.Println("update user cancelled")
		return nil
	}

	var userSelected User
	for _, user := range users {
		if user.ID == userid {
			userSelected = user
		}
	}

	var userPassword bool
	huh.NewConfirm().
		Title("Update User or Password").
		Value(&userPassword).
		Affirmative("Username").Negative("Password").
		Run()

	if userPassword {
		user1 := userSelected.Login
		user2 := userSelected.Login
		huh.NewInput().
			Title("Please enter the new username:").
			Prompt(">").
			Value(&user1).
			Run()
		huh.NewInput().
			Title("Please verify the new username:").
			Prompt(">").
			Value(&user2).
			Run()
		if user1 == "" && user2 == "" {
			return fmt.Errorf("username cannot be empty")
		}
		if user1 != user2 {
			return fmt.Errorf("usernames entered do not match")
		} else {
			userSelected.Login = user1
		}
		huh.NewConfirm().
			Title("Are you sure you want to update username for: " + userSelected.Login).
			Affirmative("yes").
			Negative("no").
			Value(&confirm).
			Run()
		if !confirm {
			return fmt.Errorf("username update cancelled")
		}
		updateStmt, err := db.Prepare("update res_users set login=$1 where id=$2;")
		if err != nil {
			fmt.Println("error preparing update statement", err)
		}
		_, err = updateStmt.Exec(userSelected.Login, userSelected.ID)
		if err != nil {
			fmt.Println("error updating user", err)
		}
		fmt.Println("update user", userSelected.Login, "successful")
	} else {
		var password1, password2 string
		huh.NewInput().
			Title("Please enter the new password:").
			Prompt(">").
			EchoMode(huh.EchoModePassword).
			Value(&password1).
			Run()
		huh.NewInput().
			Title("Please verify the new password:").
			Prompt(">").
			EchoMode(huh.EchoModePassword).
			Value(&password2).
			Run()
		if password1 == "" {
			return fmt.Errorf("password cannot be empty")
		}
		if password1 != password2 {
			return fmt.Errorf("passwords entered do not match")
		}
		huh.NewConfirm().
			Title("Are you sure you want to update password for: " + userSelected.Login).
			Affirmative("yes").
			Negative("no").
			Value(&confirm).
			Run()
		if !confirm {
			return fmt.Errorf("password update cancelled")
		}
		passkey, err := passhash.MakePassword(password1, 0, "")
		if err != nil {
			fmt.Println("password hashing error", err)
		}
		updateStmt, err := db.Prepare("update res_users set password=$1 where id=$2;")
		if err != nil {
			fmt.Println("error preparing update statement", err)
		}
		_, err = updateStmt.Exec(passkey, userSelected.ID)
		if err != nil {
			fmt.Println("error updating user", err)
		}
		fmt.Println("update user", userSelected.Login, "successful")
	}

	return nil
}

package internal

import "fmt"

func dbClone(dbhost, sourceDB, destDB, dbuser, dbpassword string) error {
	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Username: dbuser,
		Password: dbpassword,
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

	db.Exec("drop database if exists " + destDB)
	db.Exec("select pg_terminate_backend (pid) from pg_stat_activity where datname=$1", sourceDB)
	db.Exec(fmt.Sprintf("create database %s with template %s owner %s", destDB, sourceDB, dbuser))

	return nil
}

// dbReset Database Reset DBUUID and remove MCode
func dbReset(dbhost, dbname, dbuser, dbpassword string) error {
	db, err := OpenDatabase(Database{
		Hostname: dbhost,
		Username: dbuser,
		Password: dbpassword,
		Database: dbname,
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

	db.Exec("delete from ir_config_parameter where key='database.enterprise_code'")
	db.Exec(`update ir_config_parameter
		set value=(select gen_random_uuid())
		where key = 'database.uuid'`,
	)
	db.Exec(`insert into ir_config_parameter
	    (key,value,create_uid,create_date,write_uid,write_date) values
	    ('database.expiration_date',(current_date+'3 months'::interval)::timestamp,1,
	    current_timestamp,1,current_timestamp)
	    on conflict (key)
	    do update set value = (current_date+'3 months'::interval)::timestamp;`,
	)

	return nil
}

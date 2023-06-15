package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

func dsn() string {
	dbUser := parseFile("conf/odoo.conf", "db_user")
	dbPass := parseFile("conf/odoo.conf", "db_pass")
	dbHost := parseFile("conf/odoo.conf", "db_host")
	dbPort := parseFile("conf/odoo.conf", "db_port")
	dbName := parseFile("conf/odoo.conf", "db_name")
	return "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort + "/" + dbName
}

func dsnLite() string {
	dbUser := parseFile("conf/odoo.conf", "db_user")
	dbPass := parseFile("conf/odoo.conf", "db_pass")
	dbHost := parseFile("conf/odoo.conf", "db_host")
	dbPort := parseFile("conf/odoo.conf", "db_port")
	return "postgres://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort
}

var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Access the raw database",
	Long:  `Access the raw database`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// dbHost := parseConfig("db_host")
		// dbPort := parseConfig("db_port")
		// dbName := parseConfig("db_name")
		// dbUser := parseConfig("db_user")
		// dbPass := parseConfig("db_pass")
		// psql -d "host=10.0.30.90 port=5432 dbname=odoo15c user=odoo16 password=odooodoo"
		// goexecpath, _ := exec.LookPath("psql")
		// fmt.Println(goexecpath, "postgresql://"+dbUser+":"+dbPass+"@"+dbHost+":"+dbPort+"/"+dbName)
		// cmdGoVer := &exec.Cmd{
		// 	Path:   goexecpath,
		// 	Args:   []string{"postgresql://" + dbUser + ":" + dbPass + "@127.0.0.1:" + dbPort + "/" + dbName},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stdout,
		// }
		// fmt.Println(cmdGoVer.String())
		// if err := cmdGoVer.Start(); err != nil {
		// 	fmt.Println("error:", err)
		// }
		// cmdGoVer.Process.Release()
		// c := exec.Command(goexecpath, "postgresql://"+dbUser+":"+dbPass+"@"+dbHost+":"+dbPort+"/"+dbName)

		// if err := c.Start(); err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		// c.Wait()
		// fmt.Println(c.ProcessState)
		// c.Process.Release()

		// reader, err := c.StdoutPipe()
		// if err != nil {
		// 	return
		// }

		// scanner := bufio.NewScanner(reader)
		// go func() {
		// 	for scanner.Scan() {
		// 		line := scanner.Text()
		// 		fmt.Printf("%s\n", line)
		// 	}
		// }()

		// if err := cmdGoVer.Start(); err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		// if err := cmdGoVer.Wait(); err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
	},
}

func expDrop(dbName string) {
	conn, err := pgx.Connect(context.Background(), dsnLite())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	pt, err := conn.Exec(context.Background(), "DROP DATABASE IF EXISTS $1", dbName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pt)

	// def exp_drop(db_name):
	//     if db_name not in list_dbs(True):
	//         return False
	//     odoo.modules.registry.Registry.delete(db_name)
	//     odoo.sql_db.close_db(db_name)

	//     db = odoo.sql_db.db_connect('postgres')
	//     with closing(db.cursor()) as cr:
	//         # database-altering operations cannot be executed inside a transaction
	//         cr._cnx.autocommit = True
	//         _drop_conn(cr, db_name)

	//         try:
	//             cr.execute(
	//                 sql.SQL('DROP DATABASE IF EXISTS {}').format(
	//                     sql.Identifier(db_name)))
	//         except Exception as e:
	//             _logger.info('DROP DB: %s failed:\n%s', db_name, e)
	//             raise Exception("Couldn't drop database %s: %s" % (db_name, e))
	//         else:
	//             _logger.info('DROP DB: %s', db_name)

	// fs = odoo.tools.config.filestore(db_name)
	// if os.path.exists(fs):
	//
	//	shutil.rmtree(fs)
	//
	// return True
}

// SQL Dump
// c := exec.Command("pg_dump", dsn(), "--no-owner", "--file", path.Join(tPath, "dump.sql"))
// if err := c.Run(); err != nil {
// 	log.Fatal(err)
// }

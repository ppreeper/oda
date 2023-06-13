package main

import (
	"embed"
	"fmt"
	"os"
	"text/template"

	"github.com/ppreeper/oda/cmd"
)

//go:embed templates/*
var res embed.FS

func pipfile() {
	pipFile, err := template.ParseFS(res, "templates/Pipfile")
	if err != nil {
		panic(err)
	}
	fmt.Println("########## Pipfile")
	pipFile.Execute(os.Stdout, nil)
}

func envrc() {
	envrc, err := template.ParseFS(res, "templates/envrc")
	if err != nil {
		panic(err)
	}
	envrcMap := make(map[string]interface{})
	envrcMap["Version"] = "15"
	envrcMap["Port"] = "5432"
	fmt.Println("########## envrc")
	envrc.Execute(os.Stdout, envrcMap)
}

func odooconf() {
	odooConf, err := template.ParseFS(res, "templates/odoo.conf")
	if err != nil {
		panic(err)
	}
	odooConfMap := make(map[string]interface{})
	odooConfMap["Version"] = "15"
	odooConfMap["DBHost"] = "localhost"
	odooConfMap["DBPort"] = "5432"
	odooConfMap["DBName"] = "db"
	fmt.Println("########## odoo.conf")
	odooConf.Execute(os.Stdout, odooConfMap)
}

func main() {
	cmd.Execute()
	// var backup, restore, addons bool
	// flag.BoolVar(&backup, "b", false, "backup database")
	// flag.BoolVar(&restore, "r", false, "restore database")
	// flag.BoolVar(&addons, "a", false, "restore addons")
	// var dumpFile, folder string
	// flag.StringVar(&dumpFile, "d", "", "database dump file")
	// flag.StringVar(&folder, "f", "", "addons folder")
	// flag.Parse()

	// fmt.Println("backup", backup, "restore", restore, "addons", addons)
	// fmt.Println("dumpFile", dumpFile, "folder", folder)

	// dbName, err := parseConfig("db_name")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(dbName)
	// dirList, err := parseConfig("addons")
	// if err != nil {
	// 	panic(err)
	// }
	// addonDirs := strings.Split(dirList, ",")[2:]
	// fmt.Println(addonDirs)

	// if !backup && !restore && !addons {
	// 	flag.PrintDefaults()
	// 	return
	// }

	// if backup && (restore || addons) {
	// 	fmt.Println("backup or restore cannot run both commands")
	// 	return
	// }

	// if backup {
	// 	//     bkp_name = f"{time.strftime('%Y_%m_%d_%H_%M_%S')}_{db_name}"
	// 	//     print(_dump_db(db_name,bkp_name))
	// 	//     _dump_addons(addons,bkp_name)
	// 	return
	// }

	// if restore && dumpFile == "" {
	// 	fmt.Println("restore command requires a dump file to read")
	// 	return
	// }

	// if restore && dumpFile != "" {
	// 	fmt.Printf("restore from dump file %s\n", dumpFile)
	// 	//     _restore_db(db_name, args.dump_file)
	// 	return
	// }

	// if addons && dumpFile == "" {
	// 	fmt.Println("addons restore command requires a dump file to read")
	// 	return
	// }

	// if addons && dumpFile != "" {
	// 	fmt.Printf("addons restore from dump file %s\n", dumpFile)
	// 	//     _restore_addons(args.dump_file) if (args.folder is None or args.folder == "") else _restore_addons(args.dump_file,args.folder)
	// 	return
	// }
}

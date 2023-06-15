package cmd

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

//go:embed templates/*
var res embed.FS

var initprojectCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new project",
	Long:  `Create a new project`,
	Args:  cobra.MaximumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		port := int16(8069)
		dbPort := int16(5432)
		if len(args) < 1 {
			fmt.Println("invalid odoo version, valid: <15|16>")
			return
		}

		switch args[0] {
		case "15", "16":
			if len(args) < 1 {
				fmt.Println("require")
				return
			}
			version := string(args[0])
			if len(args) == 2 {
				p, err := strconv.ParseInt(args[1], 10, 16)
				if err != nil {
					fmt.Println("port must be an integer")
					return
				}
				port = int16(p)
			}
			if len(args) == 3 {
				p, err := strconv.ParseInt(args[1], 10, 16)
				if err != nil {
					fmt.Println("port must be an integer")
					return
				}
				d, err := strconv.ParseInt(args[2], 10, 16)
				if err != nil {
					fmt.Println("db_port must be an integer")
					return
				}
				port = int16(p)
				dbPort = int16(d)
			}
			initProject(version, port, dbPort)
		default:
			fmt.Println("invalid odoo version, valid: <15|16>")
		}
	},
}

func pipfile() {
	pipFile, err := template.ParseFS(res, "templates/Pipfile")
	if err != nil {
		panic(err)
	}
	file, _ := os.Create("Pipfile")
	defer file.Close()
	pipFile.Execute(file, nil)
}

func envrc(version string, port int16) {
	envrc, err := template.ParseFS(res, "templates/envrc")
	if err != nil {
		panic(err)
	}
	envrcMap := make(map[string]interface{})
	envrcMap["Version"] = version
	envrcMap["Port"] = port
	file, _ := os.Create(".envrc")
	defer file.Close()
	envrc.Execute(file, envrcMap)
}

func odooconf(version string, port, dbPort int16, dbHost, dbName string) {
	odooConf, err := template.ParseFS(res, "templates/odoo.conf")
	if err != nil {
		panic(err)
	}
	odooConfMap := make(map[string]interface{})
	odooConfMap["Version"] = version
	odooConfMap["Port"] = port
	odooConfMap["DBHost"] = dbHost
	odooConfMap["DBPort"] = dbPort
	odooConfMap["DBName"] = dbName
	file, _ := os.Create("conf/odoo.conf")
	defer file.Close()
	odooConf.Execute(file, odooConfMap)
}

func getIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	addrs, err := interfaces[1].Addrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var ip net.IP
	switch v := addrs[0].(type) {
	case *net.IPAddr:
		fmt.Println("IPAddr", v)
		ip = v.IP
	case *net.IPNet:
		fmt.Println("IPNet", v)
		ip = v.IP
	default:
		break
	}
	return ip.String()
}

func initProject(version string, port, dbPort int16) {
	paths := []string{"addons", "backups", "conf", "data"}
	epaths := []string{"data", "backups"}
	envrc(version, port)
	cmd := exec.Command("direnv", "allow")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return
	}
	for _, path := range paths {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(path, os.ModePerm)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	for _, path := range epaths {
		err := os.Chmod(path, 0o777)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	dirname, _ := os.UserHomeDir()
	repodir := dirname + "/workspace/repos/" + version + ".0/"
	err := os.Symlink(repodir+"odoo", "odoo")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.Symlink(repodir+"enterprise", "enterprise")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, cdir := getCwd()

	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}
	for _, a := range addrs {
		fmt.Println(a)
	}
	ip := getIP()

	odooconf(version, port, dbPort, ip, cdir)
	pipfile()
}

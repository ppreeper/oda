/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the instance",
	Long:    `Start the instance`,
	GroupID: "instance",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsProject() {
			return
		}
		_, project := GetProject()

		instance, err := GetInstance(project)
		if err != nil {
			fmt.Println("GetInstance", err)
			return
		}
		if instance == (Instance{}) {
			version := fmt.Sprintf("%0.1f", viper.GetFloat64("version"))
			odooimage := "odoo-" + strings.ReplaceAll(version, ".", "-")
			CopyInstance(odooimage, project)
			time.Sleep(5 * time.Second)
		}
		instance, err = GetInstance(project)
		if err != nil {
			fmt.Println("GetInstance", err)
			return
		}
		if strings.EqualFold(instance.State, "Running") {
			fmt.Fprintln(os.Stderr, project+" is already running")
			return
		}
		fmt.Println("instance not running")

		if err := IncusIdmap(project); err != nil {
			fmt.Fprintln(os.Stderr, "error idmap", err)
			return
		}

		if err := InstanceMounts(project); err != nil {
			fmt.Fprintln(os.Stderr, "InstanceMounts", err)
			return
		}

		SetInstanceState(project, "start")
		// fmt.Fprintln(os.Stderr, project, instanceStatus.Metadata.Status)

		if err := IncusHosts(project, viper.GetString("system.domain")); err != nil {
			fmt.Fprintln(os.Stderr, "error hosts", err)
			return
		}

		if err := IncusCaddyfile(project, viper.GetString("system.domain")); err != nil {
			fmt.Fprintln(os.Stderr, "error caddyfile", err)
		}

		if err := SSHConfigGenerate(project); err != nil {
			fmt.Fprintln(os.Stderr, "error sshconfig %w", err)
			return
		}
		if err = exec.Command("sshconfig").Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error sshconfig %w", err)
			return
		}

		fmt.Println(project + "." + viper.GetString("system.domain") + " started")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

// func InstanceCreate() error {
// 	if !IsProject() {
// 		return fmt.Errorf("not in a project directory")
// 	}
// 	_, project := GetProject()
// 	version := GetVersion()
// 	vers := "odoo-" + strings.Replace(version, ".", "-", -1)

// 	if err := IncusCopy(vers, project); err != nil {
// 		return fmt.Errorf("error copying %w", err)
// 	}
// 	IncusStop(project)
// 	if err := IncusIdmap(project); err != nil {
// 		return fmt.Errorf("error idmap %w", err)
// 	}
// 	return nil
// }

// func InstanceStart() error {
// 	if !IsProject() {
// 		return fmt.Errorf("not in a project directory")
// 	}
// 	conf := GetConf()
// 	_, project := GetProject()

// 	instance, err := GetInstance(project)
// 	if err != nil {
// 		InstanceCreate()
// 	}
// 	if instance.State == "RUNNING" {
// 		fmt.Println(instance.Name + " is already running")
// 		return nil
// 	}

// 	if err := InstanceMounts(project); err != nil {
// 		fmt.Println("InstanceMounts", err)
// 		return fmt.Errorf("error mounts %w", err)
// 	}

// 	if err := IncusStart(project); err != nil {
// 		fmt.Println(err)
// 		return fmt.Errorf("error starting %w", err)
// 	}

// 	if err := IncusHosts(project, GetConf().Domain); err != nil {
// 		fmt.Println(err)
// 		return fmt.Errorf("error hosts %w", err)
// 	}

// 	if err := IncusCaddyfile(project, GetConf().Domain); err != nil {
// 		fmt.Println("IncusCaddyfile", err)
// 		return fmt.Errorf("error caddyfile %w", err)
// 	}

// 	if err := SSHConfigGenerate(project); err != nil {
// 		fmt.Println("SSHConfigGenerate", err)
// 		return fmt.Errorf("error sshconfig %w", err)
// 	}
// 	if err = exec.Command("sshconfig").Run(); err != nil {
// 		fmt.Println("sshconfig", err)
// 		return fmt.Errorf("error sshconfig %w", err)
// 	}

// 	fmt.Println(project + "." + conf.Domain + " started")

// 	return nil
// }

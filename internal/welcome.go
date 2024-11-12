package internal

import (
	"os"
	"strings"

	"github.com/dimiro1/banner"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/str"
)

func cText(color, msg string) string {
	return color + msg + "{{ .AnsiColor.Default }}"
}

func (o *ODA) Welcome() error {
	tRed := "{{ .AnsiColor.BrightRed }}"
	tMagenta := "{{ .AnsiColor.Magenta }}"

	fqdn, _, _ := lib.GetFQDN()
	osName, osVersion, _ := lib.GetOSVersionName()
	osversionstring := strings.TrimSpace(osName + " " + osVersion)

	exampleDBCommands := []struct {
		Cmd  string
		Help string
	}{
		{Cmd: "oda db fullreset", Help: "Complete rebuild of the db server"},
		{Cmd: "oda db logs", Help: "Follow the postgresql logs"},
		{Cmd: "oda db psql", Help: "Direct psql access to the postgres database"},
	}
	exampleRepoCommands := []struct {
		Cmd  string
		Help string
	}{
		{Cmd: "oda repo base clone", Help: "Clone Odoo base repository"},
		{Cmd: "oda repo branch clone", Help: "Clone specific Odoo branch repository"},
		{Cmd: "oda repo base update", Help: "Update Odoo base repository"},
		{Cmd: "oda repo branch update", Help: "Update specific Odoo branch repository"},
	}
	exampleBaseCommands := []struct {
		Cmd  string
		Help string
	}{
		{Cmd: "oda base create", Help: "Create a new Odoo base image"},
		{Cmd: "oda base update", Help: "Update the Odoo base image"},
		{Cmd: "oda base destroy", Help: "Destroy the Odoo base image"},
	}

	exampleCommands := []struct {
		Cmd  string
		Help string
	}{
		{Cmd: "oda project init", Help: "Setup a new project"},
		{Cmd: "oda project reset", Help: "Drop project database and clear filestore"},
		{Cmd: "oda instance create", Help: "Initialize the project instance"},
		{Cmd: "oda start", Help: "Start the project instance"},
		{Cmd: "oda stop", Help: "Stop the project instance"},
		{Cmd: "oda restart", Help: "Restart the project instance"},
		{Cmd: "oda logs", Help: "Follow the project logs"},
		{Cmd: "oda psql", Help: "Open PostgreSQL shell of project database"},
		{Cmd: "oda exec", Help: "Open bash shell on project instance"},
		{Cmd: "oda admin updateuser", Help: "Update the an username or password"},
	}
	exampleAdminCommands := []struct {
		Cmd  string
		Help string
	}{
		{Cmd: "oda ps", Help: "List all project instances"},
		{Cmd: "oda hosts", Help: "Update the /etc/hosts file with project instances (requires sudo)"},
	}

	cmdLen := 0
	for _, cmd := range exampleDBCommands {
		if len(cmd.Cmd) > cmdLen {
			cmdLen = len(cmd.Cmd)
		}
	}
	for _, cmd := range exampleRepoCommands {
		if len(cmd.Cmd) > cmdLen {
			cmdLen = len(cmd.Cmd)
		}
	}
	for _, cmd := range exampleBaseCommands {
		if len(cmd.Cmd) > cmdLen {
			cmdLen = len(cmd.Cmd)
		}
	}
	for _, cmd := range exampleCommands {
		if len(cmd.Cmd) > cmdLen {
			cmdLen = len(cmd.Cmd)
		}
	}
	for _, cmd := range exampleAdminCommands {
		if len(cmd.Cmd) > cmdLen {
			cmdLen = len(cmd.Cmd)
		}
	}

	welcomeTemplate := cText(tMagenta, `{{ .Title "ODA" "rectangles" 0 }}`) + "\n" + cText(tMagenta, o.Version) + "\n\n"
	welcomeTemplate += "You are operating on " + cText(tRed, fqdn) + " running on " + cText(tRed, osversionstring) + "\n\n"
	welcomeTemplate += "Overview of useful commands:\n\n"

	welcomeTemplate += "database commands:\n"
	for _, cmd := range exampleDBCommands {
		welcomeTemplate += str.RJustLen("$ ", 3) + cText(tMagenta, str.LJustLen(cmd.Cmd, cmdLen+2)) + cmd.Help + "\n"
	}
	welcomeTemplate += "repository commands:\n"
	for _, cmd := range exampleRepoCommands {
		welcomeTemplate += str.RJustLen("$ ", 3) + cText(tMagenta, str.LJustLen(cmd.Cmd, cmdLen+2)) + cmd.Help + "\n"
	}
	welcomeTemplate += "base image commands:\n"
	for _, cmd := range exampleBaseCommands {
		welcomeTemplate += str.RJustLen("$ ", 3) + cText(tMagenta, str.LJustLen(cmd.Cmd, cmdLen+2)) + cmd.Help + "\n"
	}
	welcomeTemplate += "instance commands:\n"
	for _, cmd := range exampleCommands {
		welcomeTemplate += str.RJustLen("$ ", 3) + cText(tMagenta, str.LJustLen(cmd.Cmd, cmdLen+2)) + cmd.Help + "\n"
	}
	welcomeTemplate += "admin commands:\n"
	for _, cmd := range exampleAdminCommands {
		welcomeTemplate += str.RJustLen("$ ", 3) + cText(tMagenta, str.LJustLen(cmd.Cmd, cmdLen+2)) + cmd.Help + "\n"
	}
	welcomeTemplate += "\n"

	isEnabled := true
	isColorEnabled := true
	banner.InitString(os.Stdout, isEnabled, isColorEnabled, welcomeTemplate)
	return nil
}

package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/incus"
	"github.com/ppreeper/oda/lib"
	"github.com/ppreeper/oda/ui"
)

func (o *ODA) OdooBase() error {
	fmt.Println("OdooBase")
	return nil
}

func (o *ODA) OdooPS() error {
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading oda config", err)
		return nil
	}
	inc := incus.NewIncus(odaConf)

	projects := GetCurrentOdooProjects()
	instances, err := inc.GetInstances()
	if err != nil {
		fmt.Fprintln(os.Stderr, "instances list failed %w", err)
		return nil
	}

	rows := [][]string{}

	for _, instance := range instances {
		for _, project := range projects {
			if instance.Name == project {
				rows = append(rows, []string{instance.Name, instance.State, instance.IP4})
			}
		}
	}
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return ui.HeaderStyle
			case row%2 == 0:
				return ui.EvenRowStyle
			default:
				return ui.OddRowStyle
			}
		}).
		Headers("NAME", "STATE", "IPV4").
		Rows(rows...)

	fmt.Fprintln(os.Stderr, t)

	// maxnameLen := 0
	// maxstateLen := 0
	// maxipv4Len := 15

	// for _, instance := range instanceList {
	// 	if len(instance.Name) > maxnameLen {
	// 		maxnameLen = len(instance.Name)
	// 	}
	// 	if len(instance.State) > maxstateLen {
	// 		maxstateLen = len(instance.State)
	// 	}
	// }

	// fmt.Fprintf(os.Stderr, "%-*s %-*s %-*s\n",
	// 	maxnameLen+2, "NAME",
	// 	maxstateLen+2, "STATE",
	// 	maxipv4Len+2, "IPV4",
	// )
	// for _, instance := range instanceList {
	// 	fmt.Fprintf(os.Stderr, "%-*s %-*s %-*s\n",
	// 		maxnameLen+2, instance.Name,
	// 		maxstateLen+2, instance.State,
	// 		maxipv4Len+2, instance.IP4,
	// 	)
	// }
	return nil
}

// OdooPSQL
// calls into "incus exec <project> -t -- pgsql"
func (o *ODA) OdooPSQL() error {
	if !IsProject() {
		return nil
	}
	cwd, _ := lib.GetProject()
	odooConf, _ := config.LoadOdooConfig(cwd)
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

	dbHost := odooConf.DbHost
	dbuser := odooConf.DbUser
	dbpassword := odooConf.DbPassword
	dbname := odooConf.DbName

	uid, err := inc.IncusGetUid(dbHost, "postgres")
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("could not get postgres uid %v"), err)
		return nil
	}

	incusCmd := exec.Command("incus", "exec", dbHost, "--user", uid,
		"--env", "PGPASSWORD="+dbpassword, "-t", "--",
		"psql", "-h", "127.0.0.1", "-U", dbuser, dbname,
	)
	incusCmd.Stdin = os.Stdin
	incusCmd.Stdout = os.Stdout
	incusCmd.Stderr = os.Stderr
	if err := incusCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error instance psql %w", err)
	}
	return nil
}

func (o *ODA) OdooCreate() error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()
	projectConfig, _ := config.LoadProjectConfig()

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

	version := projectConfig.Version
	verParts := strings.Split(version, ".")
	if len(verParts) < 1 {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("invalid version %s in .oda.yaml", version))
		return nil
	}
	baseVersion := "odoo-" + verParts[0] + "-0"

	iStatus := inc.GetInstanceState(project)
	if iStatus.StatusCode == 200 {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render(project, "already exists"))
		return nil
	}

	inc.CopyInstance(baseVersion, project)

	time.Sleep(5 * time.Second)

	if err := inc.IncusIdmap(project); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("error idmap %v"), err)
		return nil
	}

	if err := inc.InstanceMounts(project); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("InstanceMounts %v"), err)
		return nil
	}

	return nil
}

// OdooDestroy
// full instance destroy
// stop instance
// remove instance
func (o *ODA) OdooDestroy() error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()
	confim := ui.AreYouSure("destroy the " + project + " instance")
	if !confim {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("destroying the "+project+" instance canceled"))
		return nil
	}

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading oda config", err)
		return nil
	}
	inc := incus.NewIncus(odaConf)

	inc.SetInstanceState(project, "stop")
	inc.DeleteInstance(project)
	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("instance", project, "destroyed"))
	return nil
}

// OdooExec
// execCmd.Flags().StringVarP(&username, "username", "u", "odoo", "username")
// calls into "incus exec <project> -t -- /bin/bash"
func (o *ODA) OdooExec(username string) error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

	iStatus := inc.GetInstanceState(project)
	if iStatus.Metadata.Status != "Running" {
		fmt.Fprintln(os.Stderr, ui.WarningStyle.Render(project, "stopped, please start"))
		return nil
	}

	uid, err := inc.IncusGetUid(project, username)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("could not get odoo uid %v"), err)
		return nil
	}

	incusCmd := exec.Command("incus", "exec", project, "--user", uid, "-t",
		"--env", "HOME=/home/odoo", "--cwd", "/opt/odoo", "--",
		"/bin/bash",
	)
	incusCmd.Stdin = os.Stdin
	incusCmd.Stdout = os.Stdout
	incusCmd.Stderr = os.Stderr
	if err := incusCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("error executing %v"), err)
		return nil
	}
	return nil
}

// OdooStart
// full instance start
// incus start <project>
func (o *ODA) OdooStart() error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading oda config", err)
		return nil
	}
	inc := incus.NewIncus(odaConf)

	instance, err := inc.GetInstance(project)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("could not get odoo instance"), err)
		return nil
	}
	if instance == (incus.Instance{}) {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("no odoo instance, please create one first"))
		return nil
	}

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Starting", project))
	switch strings.ToUpper(instance.State) {
	case "STOPPED":
		inc.SetInstanceState(project, "start")
	}

	instanceStatus := inc.GetInstanceState(project)
	fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(project, instanceStatus.Metadata.Status))

	if err := inc.IncusExec(project, "sudo", "odas", "hosts", odaConf.System.Domain); err != nil {
		fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("error hosts %v\n"), err)
		return nil
	}

	if err := inc.IncusExec(project, "sudo", "odas", "caddy", odaConf.System.Domain); err != nil {
		fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("error caddyfile %v\n"), err)
		return nil
	}

	if err := SSHConfigGenerate(project); err != nil {
		fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("error sshconfig %v\n"), err)
		return nil
	}
	if err = exec.Command("sshconfig").Run(); err != nil {
		fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("error sshconfig %v\n"), err)
		return nil
	}

	return nil
}

// OdooStop
// full instance stop
// incus stop <project>
func (o *ODA) OdooStop() error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()

	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Stopping", project))
	inc.SetInstanceState(project, "stop")
	instanceStatus := inc.GetInstanceState(project)
	fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(project, instanceStatus.Metadata.Status))
	return nil
}

// OdooRestart
// full instance restart
// incus restart <project>
func (o *ODA) OdooRestart() error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()

	odaConf, _ := config.LoadOdaConfig()
	inc := incus.NewIncus(odaConf)

	instance, err := inc.GetInstance(project)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("could not get odoo instance"), err)
		return nil
	}
	if instance == (incus.Instance{}) {
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("no odoo instance, please launch one first"))
		return nil
	}

	fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Restarting", project))
	switch strings.ToUpper(instance.State) {
	case "RUNNING":
		inc.SetInstanceState(project, "restart")
	case "STOPPED":
		inc.SetInstanceState(project, "start")
	}

	instanceStatus := inc.GetInstanceState(project)
	fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render(project, instanceStatus.Metadata.Status))
	return nil
}

// OdooLogs
// calls into "odas logs"
// calls into "incus exec <project> -t -- journalctl -f"
func (o *ODA) OdooLogs() error {
	if !IsProject() {
		return nil
	}
	_, project := lib.GetProject()
	podCmd := exec.Command("incus",
		"exec", project, "-t", "--",
		"journalctl", "-f",
	)
	podCmd.Stdin = os.Stdin
	podCmd.Stdout = os.Stdout
	podCmd.Stderr = os.Stderr
	if err := podCmd.Run(); err != nil {
		// fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("error getting logs %v"), err)
		return nil
	}
	return nil
}

func SSHConfigGenerate(project string) error {
	HOME, _ := os.UserHomeDir()
	// conf := GetConf()
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return fmt.Errorf("load oda config failed %w", err)
	}
	inc := incus.NewIncus(odaConf)
	domain := odaConf.System.Domain
	sshkey := odaConf.System.SSHKey
	instance, err := inc.GetInstance(project)
	if err != nil {
		return fmt.Errorf("error getting instance %w", err)
	}
	// priority;host;hostname;user;identityfile;port
	sshconfig := fmt.Sprintf("%d;%s.%s;%s;%s;%s;%d", 10, project, domain, instance.IP4, "odoo", sshkey, 22)
	sshconfigCSV := filepath.Join(HOME, ".ssh", "sshconfig.csv")
	// READ config
	hosts, err := os.Open(sshconfigCSV)
	if err != nil {
		return fmt.Errorf("hosts file read failed %w", err)
	}
	defer hosts.Close()
	hostlines := []string{}
	scanner := bufio.NewScanner(hosts)
	for scanner.Scan() {
		hostlines = append(hostlines, scanner.Text())
	}
	headerLine := hostlines[0]
	// Remove old lines
	newHostlines := []string{}
	for _, hostline := range hostlines[1:] {
		hostlineSplit := strings.Split(hostline, ";")
		if hostlineSplit[1] == project+"."+domain {
			continue
		}
		newHostlines = append(newHostlines, hostline)
	}
	// Add new lines
	newHostlines = append(newHostlines, sshconfig)
	// WRITE config
	fo, err := os.Create(sshconfigCSV)
	if err != nil {
		return fmt.Errorf("hosts file write failed %w", err)
	}
	defer fo.Close()
	fo.WriteString(headerLine + "\n")
	for _, hostline := range newHostlines {
		fo.WriteString(hostline + "\n")
	}

	return nil
}

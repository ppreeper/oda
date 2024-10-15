package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

func moduleList(modules ...string) string {
	mods := []string{}
	for _, mod := range modules {
		mm := strings.Split(mod, ",")
		mods = append(mods, mm...)
	}
	return strings.Join(removeDuplicate(mods), ",")
}

// GetCurrentOdooProjects Get Current Odoo Projects
func GetOdooBackups(project string) (backups, addons []string) {
	// dirs := GetDirs()
	viper.GetString("dirs.project")
	entries, err := os.ReadDir(filepath.Join(viper.GetString("dirs.project"), "backups"))
	if err != nil {
		fmt.Println(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			fname := strings.Split(entry.Name(), "__")
			if len(fname) == 2 {
				backups = append(backups, entry.Name())
			} else if len(fname) == 3 {
				addons = append(addons, entry.Name())
			}
		}
	}
	slices.Sort(backups)
	slices.Sort(addons)
	backups = removeDuplicate(backups)
	addons = removeDuplicate(addons)
	if project != "" {
		backups = selectOnly(backups, project)
		addons = selectOnly(addons, project)
	}
	return
}

func GetOdooConf(cwd, key string) string {
	// cwd, _ := GetProject()
	odooconf := filepath.Join(cwd, "conf", "odoo.conf")
	c, err := os.Open(odooconf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading odoo.conf file", err)
		return ""
	}
	defer func() {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}()
	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		line := scanner.Text()
		re := regexp.MustCompile(`^` + key + ` = (.+)$`)
		if re.MatchString(line) {
			match := re.FindStringSubmatch(line)
			return match[1]
		}
	}
	return ""
}

// GetCurrentOdooRepos Get Currently Copied Odoo Repos
func GetCurrentOdooRepos() []string {
	repoDir := viper.GetString("dirs.repo")

	dirnames := []string{}
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirnames = append(dirnames, entry.Name())
		}
	}
	slices.Sort(dirnames)
	slices.Reverse(dirnames)
	dirnames = removeDuplicate(dirnames)
	dirnames = removeValue(dirnames, "odoo")
	dirnames = removeValue(dirnames, "enterprise")
	dirnames = removeValue(dirnames, "design-themes")
	dirnames = removeValue(dirnames, "industry")
	return dirnames
}

// GetCurrentOdooProjects Get Current Odoo Projects
func GetCurrentOdooProjects() []string {
	projectDir := viper.GetString("dirs.project")

	dirnames := []string{}
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirnames = append(dirnames, entry.Name())
		}
	}
	slices.Sort(dirnames)
	dirnames = removeDuplicate(dirnames)
	dirnames = removeValue(dirnames, "backups")
	dirnames = removeValue(dirnames, "odoo")
	dirnames = removeValue(dirnames, "enterprise")
	return dirnames
}

func GetProject() (string, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	return cwd, filepath.Base(cwd)
}

func IsProject() bool {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	base := filepath.Base(cwd)
	pdir := filepath.Join(viper.GetString("dirs.project"), base)
	odooconf := filepath.Join(cwd, "conf", "odoo.conf")

	if cwd != pdir {
		fmt.Fprintln(os.Stderr, "not in a project directory")
		return false
	}
	if _, err := os.Stat(odooconf); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "not in a project directory")
		return false
	}
	return true
}

func GetGitHubUsernameToken() (username, token string) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fi, err := os.Open(filepath.Join(homedir, ".gitcreds"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	scanner := bufio.NewScanner(fi)
	re := regexp.MustCompile(`https://(.+):(.+)@github.com$`)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			match := re.FindStringSubmatch(line)
			username = match[1]
			token = match[2]
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("error opening directory %v", err)
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("error reading directory names %v", err)
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("error removing directory %v", err)
		}
	}
	return nil
}

func AreYouSure(prompt string) bool {
	var confirm1, confirm2 bool
	huh.NewConfirm().
		Title(fmt.Sprintf("Are you sure you want to %s?", prompt)).
		Value(&confirm1).
		Run()
	if !confirm1 {
		return false
	}
	huh.NewConfirm().
		Title(fmt.Sprintf("Are you really sure you want to %s?", prompt)).
		Value(&confirm2).
		Run()
	if !confirm1 || !confirm2 {
		return false
	}
	return true
}

// existsIn searches list for value
func existsIn[T comparable](sliceList []T, value T) bool {
	for _, item := range sliceList {
		if value == item {
			return true
		}
	}
	return false
}

// removeDuplicate Remove Duplicate Values from Slice
func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// removeValue Remove Value from Slice
func removeValue[T comparable](slice []T, value T) []T {
	for i, item := range slice {
		if item == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func selectOnly(sliceList []string, value string) []string {
	list := []string{}
	for _, item := range sliceList {
		if strings.Contains(item, value) {
			list = append(list, item)
		}
	}
	return list
}

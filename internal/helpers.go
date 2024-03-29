package oda

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/joho/godotenv"
)

type Dirs struct {
	Repo    string
	Project string
}

func GetDirs() Dirs {
	sudouser, _ := os.LookupEnv("SUDO_USER")
	DirList := Dirs{}
	var CONFIG string
	if sudouser != "" {
		userLookup, _ := user.Lookup(sudouser)
		CONFIG = filepath.Join(userLookup.HomeDir, ".config")
	} else {
		CONFIG, _ = os.UserConfigDir()
	}
	ODOOCONF := filepath.Join(CONFIG, "oda", "oda.conf")
	err := godotenv.Load(ODOOCONF)
	if err != nil {
		fmt.Println("Error loading .env file")
		return DirList
	}
	DirList.Repo, _ = os.LookupEnv("REPO")
	DirList.Project, _ = os.LookupEnv("PROJECT")
	return DirList
}

type Conf struct {
	Odoobase   string
	Domain     string
	OSImage    string
	DBImage    string
	BRAddr     string
	DBH        string
	DBHost     string
	DBPort     string
	DBPass     string
	DBUsername string
	DBUserpass string
}

func GetConf() Conf {
	sudouser, _ := os.LookupEnv("SUDO_USER")
	ConfList := Conf{}
	var CONFIG string
	if sudouser != "" {
		userLookup, _ := user.Lookup(sudouser)
		CONFIG = filepath.Join(userLookup.HomeDir, ".config")
	} else {
		CONFIG, _ = os.UserConfigDir()
	}
	ODOOCONF := filepath.Join(CONFIG, "oda", "oda.conf")
	err := godotenv.Load(ODOOCONF)
	if err != nil {
		fmt.Println("Error loading .env file")
		return ConfList
	}
	ConfList.Odoobase, _ = os.LookupEnv("ODOOBASE")
	ConfList.Domain, _ = os.LookupEnv("DOMAIN")
	ConfList.OSImage, _ = os.LookupEnv("OS_IMAGE")
	ConfList.DBImage, _ = os.LookupEnv("DB_IMAGE")
	ConfList.BRAddr, _ = os.LookupEnv("BR_ADDR")
	ConfList.DBH, _ = os.LookupEnv("DBH")
	ConfList.DBHost, _ = os.LookupEnv("DB_HOST")
	ConfList.DBPort, _ = os.LookupEnv("DB_PORT")
	ConfList.DBPass, _ = os.LookupEnv("DB_PASS")
	ConfList.DBUsername, _ = os.LookupEnv("DB_USERNAME")
	ConfList.DBUserpass, _ = os.LookupEnv("DB_USERPASS")
	return ConfList
}

func GetOdooConf(cwd, key string) string {
	// cwd, _ := GetProject()
	odooconf := filepath.Join(cwd, "conf", "odoo.conf")
	c, err := os.Open(odooconf)
	if err != nil {
		fmt.Println("Error loading odoo.conf file", err)
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

func GetVersion() string {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
		return ""
	}
	version, exists := os.LookupEnv("ODOO_V")
	if !exists {
		fmt.Println("ODOO_V not set")
		return ""
	}
	return version
}

func GetProject() (string, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", ""
	}
	return cwd, filepath.Base(cwd)
}

func IsProject() bool {
	dirLst := GetDirs()
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	base := filepath.Base(cwd)
	pdir := filepath.Join(dirLst.Project, base)
	odooconf := filepath.Join(cwd, "conf", "odoo.conf")

	if cwd != pdir {
		fmt.Println("not in a project directory")
		return false
	}
	if _, err := os.Stat(odooconf); os.IsNotExist(err) {
		fmt.Println("not in a project directory")
		return false
	}
	return true
}

func GetGitHubUsernameToken() (username, token string) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	fi, err := os.Open(filepath.Join(homedir, ".gitcreds"))
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
	}
	return
}

func CloneUrlDir(url, baseDir, cloneDir, username, token string) error {
	_, err := os.Stat(filepath.Join(baseDir, cloneDir, ".git"))
	if os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0o755)
		_, err = git.PlainClone(filepath.Join(baseDir, cloneDir), false, &git.CloneOptions{
			URL:      url,
			Progress: os.Stdout,
			Auth: &http.BasicAuth{
				Username: username,
				Password: token,
			},
		})
		return err
	}
	return nil
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
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

// GetCurrentOdooRepos Get Currently Copied Odoo Repos
func GetCurrentOdooRepos() []string {
	dirnames := []string{}
	dirs := GetDirs()
	entries, err := os.ReadDir(dirs.Repo)
	if err != nil {
		fmt.Println(err)
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
	return dirnames
}

// GetCurrentOdooProjects Get Current Odoo Projects
func GetCurrentOdooProjects() []string {
	dirnames := []string{}
	dirs := GetDirs()
	entries, err := os.ReadDir(dirs.Project)
	if err != nil {
		fmt.Println(err)
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

func GetOdooBackupsNode() (backups, addons []string) {
	entries, err := os.ReadDir(filepath.Join("/opt/odoo", "backups"))
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
	return
}

// GetCurrentOdooProjects Get Current Odoo Projects
func GetOdooBackups() (backups, addons []string) {
	dirs := GetDirs()
	entries, err := os.ReadDir(filepath.Join(dirs.Project, "backups"))
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
	return
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

// existsIn searches list for value
func existsIn[T comparable](sliceList []T, value T) bool {
	for _, item := range sliceList {
		if value == item {
			return true
		}
	}
	return false
}

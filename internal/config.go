package oda

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Conf struct {
	Repo       string `json:"REPO,omitempty"`
	Project    string `json:"PROJECT,omitempty"`
	Domain     string `json:"DOMAIN,omitempty"`
	OSImage    string `json:"OS_IMAGE,omitempty"`
	DBImage    string `json:"DB_IMAGE,omitempty"`
	DBHost     string `json:"DB_HOST,omitempty"`
	DBPort     string `json:"DB_PORT,omitempty"`
	DBPass     string `json:"DB_PASS,omitempty"`
	DBUsername string `json:"DB_USERNAME,omitempty"`
	DBUserpass string `json:"DB_USERPASS,omitempty"`
	SSHKey     string `json:"SSH_KEY,omitempty"`
}

func (o *Conf) WithRepo(repodir string) *Conf {
	o.Repo = repodir
	return o
}

func (o *Conf) WithProject(projectdir string) *Conf {
	o.Project = projectdir
	return o
}

func (o *Conf) WithDomain(domain string) *Conf {
	o.Domain = domain
	return o
}

func (o *Conf) WithOSImage(osimage string) *Conf {
	o.OSImage = osimage
	return o
}

func (o *Conf) WithDBImage(dbimage string) *Conf {
	o.DBImage = dbimage
	return o
}

func (o *Conf) WithDBHost(dbhost string) *Conf {
	o.DBHost = dbhost
	return o
}

func (o *Conf) WithDBPort(dbport string) *Conf {
	o.DBPort = dbport
	return o
}

func (o *Conf) WithDBPass(dbpass string) *Conf {
	o.DBPass = dbpass
	return o
}

func (o *Conf) WithDBUsername(dbusername string) *Conf {
	o.DBUsername = dbusername
	return o
}

func (o *Conf) WithDBUserpass(dbuserpass string) *Conf {
	o.DBUserpass = dbuserpass
	return o
}

func NewConf() *Conf {
	return &Conf{
		Domain:     "local",
		OSImage:    "ubuntu/22.04",
		DBImage:    "debian/12",
		DBHost:     "db",
		DBPort:     "6432",
		DBPass:     "postgres",
		DBUsername: "odoodev",
		DBUserpass: "odooodoo",
		SSHKey:     "id_rsa",
	}
}

func (c *Conf) GetURI() {
	c.DBHost = fmt.Sprintf("db.%s", c.Domain)
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
	ConfList.Repo, _ = os.LookupEnv("REPO")
	ConfList.Project, _ = os.LookupEnv("PROJECT")
	ConfList.Domain, _ = os.LookupEnv("DOMAIN")
	ConfList.OSImage, _ = os.LookupEnv("OS_IMAGE")
	ConfList.DBImage, _ = os.LookupEnv("DB_IMAGE")
	ConfList.DBHost, _ = os.LookupEnv("DB_HOST")
	ConfList.DBPort, _ = os.LookupEnv("DB_PORT")
	ConfList.DBPass, _ = os.LookupEnv("DB_PASS")
	ConfList.DBUsername, _ = os.LookupEnv("DB_USERNAME")
	ConfList.DBUserpass, _ = os.LookupEnv("DB_USERPASS")
	ConfList.SSHKey, _ = os.LookupEnv("SSH_KEY")
	return ConfList
}

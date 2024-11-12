package config

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type OdooConfig struct {
	AddonsPath        []string `yaml:"addons_path"`
	DataDir           string   `yaml:"data_dir"`
	AdminPasswd       string   `yaml:"admin_passwd"`
	WithoutDemo       string   `yaml:"without_demo"`
	CsvInternalSep    string   `yaml:"csv_internal_sep"`
	Reportgz          string   `yaml:"reportgz"`
	ServerWideModules string   `yaml:"server_wide_modules"`
	DbHost            string   `yaml:"db_host"`
	DbPort            string   `yaml:"db_port"`
	DbMaxconn         int      `yaml:"db_maxconn"`
	DbUser            string   `yaml:"db_user"`
	DbPassword        string   `yaml:"db_password"`
	DbName            string   `yaml:"db_name"`
	DbTemplate        string   `yaml:"db_template"`
	DbSslmode         string   `yaml:"db_sslmode"`
	ListDb            bool     `yaml:"list_db"`
	Proxy             bool     `yaml:"proxy"`
	ProxyMode         bool     `yaml:"proxy_mode"`
	Logfile           string   `yaml:"logfile"`
	LogLevel          string   `yaml:"log_level"`
	LogHandler        string   `yaml:"log_handler"`
	Workers           int      `yaml:"workers"`
}

func NewOdooConfig() *OdooConfig {
	return &OdooConfig{
		AddonsPath:        []string{"/opt/odoo/odoo/addons", "/opt/odoo/design-themes", "/opt/odoo/industry", "/opt/odoo/addons"},
		DataDir:           "/opt/odoo/data",
		AdminPasswd:       "adminadmin",
		WithoutDemo:       "all",
		CsvInternalSep:    ";",
		Reportgz:          "false",
		ServerWideModules: "base,web",
		DbHost:            "localhost",
		DbPort:            "5432",
		DbMaxconn:         8,
		DbUser:            "odoo",
		DbPassword:        "odoo",
		DbName:            "odoo",
		DbTemplate:        "template0",
		DbSslmode:         "disable",
		ListDb:            false,
		Proxy:             true,
		ProxyMode:         true,
		Logfile:           "/dev/stderr",
		LogLevel:          "debug",
		LogHandler:        "odoo.tools.convert:DEBUG",
		Workers:           2,
	}
}

func (odoo *OdooConfig) Write(projectName, projectDir, edition string, embedFS embed.FS) error {
	odaConf, err := LoadOdaConfig()
	if err != nil {
		return err
	}

	projectName = strings.ReplaceAll(projectName, "-", "_")
	dbname := projectName + "_" + odaConf.System.Domain

	odooConfFile := filepath.Join(projectDir, "conf", "odoo.conf")

	fo, err := os.Create(odooConfFile)
	if err != nil {
		return fmt.Errorf("cannot create odoo.conf file %w", err)
	}
	defer fo.Close()

	data := map[string]string{
		"enterprise_dir":      "",
		"admin_passwd":        odoo.AdminPasswd,
		"without_demo":        odoo.WithoutDemo,
		"reportgz":            odoo.Reportgz,
		"server_wide_modules": odoo.ServerWideModules,
		"db_host":             odaConf.Database.Host,
		"db_port":             fmt.Sprintf("%d", odaConf.Database.Port),
		"db_maxconn":          fmt.Sprintf("%d", odoo.DbMaxconn),
		"db_user":             odaConf.Database.Username,
		"db_password":         odaConf.Database.Password,
		"db_name":             dbname,
		"db_template":         odoo.DbTemplate,
		"db_sslmode":          odoo.DbSslmode,
		"list_db":             fmt.Sprintf("%t", odoo.ListDb),
		"proxy":               fmt.Sprintf("%t", odoo.Proxy),
		"proxy_mode":          fmt.Sprintf("%t", odoo.ProxyMode),
		"logfile":             odoo.Logfile,
		"log_level":           odoo.LogLevel,
		"log_handler":         odoo.LogHandler,
		"workers":             fmt.Sprintf("%d", odoo.Workers),
	}
	if edition == "enterprise" {
		data["enterprise_dir"] = "/opt/odoo/enterprise,"
	}

	// load and write template
	t, err := template.ParseFS(embedFS, "templates/odoo.conf")
	if err != nil {
		return fmt.Errorf("cannot parse odoo.conf template %w", err)
	}
	err = t.Execute(fo, data)
	if err != nil {
		return fmt.Errorf("cannot write odoo.conf file %w", err)
	}

	return nil
}

func LoadOdooConfig(cwd string) (*OdooConfig, error) {
	// fmt.Println("LoadOdooConfig", cwd)

	confFileName := filepath.Join(cwd, "conf", "odoo.conf")
	// fmt.Println("confFile", confFileName)

	confFile, err := os.Open(confFileName)
	if err != nil {
		return nil, fmt.Errorf("cannot open odoo.conf file %w", err)
	}
	defer func() {
		if err := confFile.Close(); err != nil {
			panic(err)
		}
	}()

	kvals := map[string]string{}

	scanner := bufio.NewScanner(confFile)
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.SplitN(line, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			kvals[key] = value
		}
	}
	// fmt.Println(kvals)

	odooConf := BuildOdooConfig(kvals)
	// fmt.Println("Loaded:", odooConf)

	// fmt.Println("LoadOdaConfig", odaConf)
	return odooConf, nil
}

func BuildOdooConfig(kvals map[string]string) *OdooConfig {
	odooConf := OdooConfig{}
	t := reflect.TypeOf(odooConf)
	if t.Kind() != reflect.Struct {
		fmt.Println("not a struct")
		return &odooConf
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if tag == "" {
			continue
		}

		if val, ok := kvals[tag]; ok {
			// reflect.ValueOf(&odooConf).Elem().Field(i).SetString(val)
			// fmt.Printf("%d. %v (%v), tag: '%v', val: %s\n", i+1, field.Name, field.Type.Name(), tag, val)
			switch field.Type.Name() {
			case "string":
				reflect.ValueOf(&odooConf).Elem().Field(i).SetString(val)
			case "int":
				iVal, err := strconv.Atoi(val)
				if err != nil {
					fmt.Println("error converting to int", err)
				}
				reflect.ValueOf(&odooConf).Elem().Field(i).SetInt(int64(iVal))
			case "bool":
				reflect.ValueOf(&odooConf).Elem().Field(i).SetBool(val == "true")
			}
		}

		// if _, ok := t.Field(i).Tag.Lookup("json"); ok {
		// 	fmt.Println("tag", t.Field(i).Tag.Get("yaml"))
		// }
	}
	return &odooConf
}

func ReadConfValue(conffile, key, def string) string {
	c, err := os.Open(conffile)
	if err != nil {
		return def
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
	return def
}

// writeOdooConf Write Odoo Configfile
func WriteOdooConf(file, projectName, edition string, embedFS embed.FS) error {
	// fmt.Println("writeOdooConf", file, projectName, edition)
	odaConf, err := LoadOdaConfig()
	if err != nil {
		return err
	}

	projectName = strings.ReplaceAll(projectName, "-", "_")
	dbname := projectName + "_" + odaConf.System.Domain

	fo, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("cannot create odoo.conf file %w", err)
	}
	defer fo.Close()

	data := map[string]string{
		"db_host":        odaConf.Database.Host,
		"db_port":        string(odaConf.Database.Port),
		"db_user":        odaConf.Database.Username,
		"db_password":    odaConf.Database.Password,
		"db_name":        dbname,
		"enterprise_dir": "",
	}
	if edition == "enterprise" {
		data["enterprise_dir"] = "/opt/odoo/enterprise,"
	}
	t, err := template.ParseFS(embedFS, "templates/odoo.conf")
	if err != nil {
		return fmt.Errorf("cannot parse odoo.conf template %w", err)
	}
	err = t.Execute(fo, data)
	if err != nil {
		return fmt.Errorf("cannot write odoo.conf file %w", err)
	}

	return nil
}

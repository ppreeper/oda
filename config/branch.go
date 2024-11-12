package config

import (
	"golang.org/x/exp/slices"
)

const OdooBaseURL = "https://github.com/odoo"

type Branch struct {
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	Image            string   `json:"image"`
	InstanceName     string   `json:"instance_name"`
	Repos            []string `json:"repos"`
	BaselinePackages []string `json:"baseline_packages"`
	Odoobase         []string `json:"odoobase"`
}

func GetBranches() []*Branch {
	return []*Branch{
		{
			Name:         "15.0",
			Version:      "15.0",
			Image:        "ubuntu/22.04",
			InstanceName: "odoo-15-0",
			Repos:        []string{"odoo", "enterprise", "design-themes"},
			BaselinePackages: []string{
				"apt-transport-https", "apt-utils", "bzip2", "ca-certificates", "curl",
				"dirmngr", "git", "gnupg", "inetutils-ping", "libgnutls-dane0", "libgts-bin",
				"libpaper-utils", "locales", "lsb-release", "nodejs", "npm", "odaserver",
				"openssh-server", "postgresql-common", "python3", "python3-full",
				"shared-mime-info", "sudo", "unzip", "vim", "wget", "xz-utils", "zip", "zstd",
			},
			Odoobase: []string{
				"fonts-liberation", "fonts-noto", "fonts-noto-cjk", "fonts-noto-mono",
				"geoip-database", "gsfonts", "python3-babel", "python3-chardet",
				"python3-cryptography", "python3-cups", "python3-dateutil",
				"python3-decorator", "python3-docutils", "python3-feedparser",
				"python3-freezegun", "python3-geoip2", "python3-gevent", "python3-googleapi",
				"python3-greenlet", "python3-html2text", "python3-idna", "python3-jinja2",
				"python3-ldap", "python3-libsass", "python3-lxml", "python3-markupsafe",
				"python3-num2words", "python3-odf", "python3-ofxparse", "python3-olefile",
				"python3-openssl", "python3-paramiko", "python3-passlib", "python3-pdfminer",
				"python3-phonenumbers", "python3-pil", "python3-pip", "python3-polib",
				"python3-psutil", "python3-psycopg2", "python3-pydot", "python3-pylibdmtx",
				"python3-pyparsing", "python3-pypdf2", "python3-qrcode", "python3-renderpm",
				"python3-reportlab", "python3-reportlab-accel", "python3-requests",
				"python3-rjsmin", "python3-serial", "python3-setuptools", "python3-stdnum",
				"python3-tz", "python3-urllib3", "python3-usb", "python3-vobject",
				"python3-werkzeug", "python3-xlrd", "python3-xlsxwriter", "python3-xlwt",
				"python3-zeep",
			},
		},
		{
			Name:         "16.0",
			Version:      "16.0",
			Image:        "ubuntu/22.04",
			InstanceName: "odoo-16-0",
			Repos:        []string{"odoo", "enterprise", "design-themes", "industry"},
			BaselinePackages: []string{
				"apt-transport-https", "apt-utils", "bzip2", "ca-certificates", "curl",
				"dirmngr", "git", "gnupg", "inetutils-ping", "libgnutls-dane0", "libgts-bin",
				"libpaper-utils", "locales", "lsb-release", "nodejs", "npm", "odaserver",
				"openssh-server", "postgresql-common", "python3", "python3-full",
				"shared-mime-info", "sudo", "unzip", "vim", "wget", "xz-utils", "zip", "zstd",
			},
			Odoobase: []string{
				"fonts-liberation", "fonts-noto", "fonts-noto-cjk", "fonts-noto-mono",
				"geoip-database", "gsfonts", "python3-babel", "python3-chardet",
				"python3-cryptography", "python3-cups", "python3-dateutil",
				"python3-decorator", "python3-docutils", "python3-feedparser",
				"python3-freezegun", "python3-geoip2", "python3-gevent", "python3-googleapi",
				"python3-greenlet", "python3-html2text", "python3-idna", "python3-jinja2",
				"python3-ldap", "python3-libsass", "python3-lxml", "python3-markupsafe",
				"python3-num2words", "python3-odf", "python3-ofxparse", "python3-olefile",
				"python3-openssl", "python3-paramiko", "python3-passlib", "python3-pdfminer",
				"python3-phonenumbers", "python3-pil", "python3-pip", "python3-polib",
				"python3-psutil", "python3-psycopg2", "python3-pydot", "python3-pylibdmtx",
				"python3-pyparsing", "python3-pypdf2", "python3-qrcode", "python3-renderpm",
				"python3-reportlab", "python3-reportlab-accel", "python3-requests",
				"python3-rjsmin", "python3-serial", "python3-setuptools", "python3-stdnum",
				"python3-tz", "python3-urllib3", "python3-usb", "python3-vobject",
				"python3-werkzeug", "python3-xlrd", "python3-xlsxwriter", "python3-xlwt",
				"python3-zeep",
			},
		},
		{
			Name:         "17.0",
			Version:      "17.0",
			Image:        "ubuntu/22.04",
			InstanceName: "odoo-17-0",
			Repos:        []string{"odoo", "enterprise", "design-themes", "industry"},
			BaselinePackages: []string{
				"apt-transport-https", "apt-utils", "bzip2", "ca-certificates", "curl",
				"dirmngr", "git", "gnupg", "inetutils-ping", "libgnutls-dane0", "libgts-bin",
				"libpaper-utils", "locales", "lsb-release", "nodejs", "npm", "odaserver",
				"openssh-server", "postgresql-common", "python3", "python3-full",
				"shared-mime-info", "sudo", "unzip", "vim", "wget", "xz-utils", "zip", "zstd",
			},
			Odoobase: []string{
				"fonts-liberation", "fonts-noto", "fonts-noto-cjk", "fonts-noto-mono",
				"geoip-database", "gsfonts", "python3-babel", "python3-chardet",
				"python3-cryptography", "python3-cups", "python3-dateutil",
				"python3-decorator", "python3-docutils", "python3-feedparser",
				"python3-freezegun", "python3-geoip2", "python3-gevent", "python3-googleapi",
				"python3-greenlet", "python3-html2text", "python3-idna", "python3-jinja2",
				"python3-ldap", "python3-libsass", "python3-lxml", "python3-markupsafe",
				"python3-num2words", "python3-odf", "python3-ofxparse", "python3-olefile",
				"python3-openssl", "python3-paramiko", "python3-passlib", "python3-pdfminer",
				"python3-phonenumbers", "python3-pil", "python3-pip", "python3-polib",
				"python3-psutil", "python3-psycopg2", "python3-pydot", "python3-pylibdmtx",
				"python3-pyparsing", "python3-pypdf2", "python3-qrcode", "python3-renderpm",
				"python3-reportlab", "python3-reportlab-accel", "python3-requests",
				"python3-rjsmin", "python3-serial", "python3-setuptools", "python3-stdnum",
				"python3-tz", "python3-urllib3", "python3-usb", "python3-vobject",
				"python3-werkzeug", "python3-xlrd", "python3-xlsxwriter", "python3-xlwt",
				"python3-zeep",
			},
		},
		{
			Name:             "saas-17.2",
			Version:          "17.2",
			Image:            "ubuntu/22.04",
			InstanceName:     "odoo-17-0",
			Repos:            []string{"odoo", "enterprise", "design-themes", "industry"},
			BaselinePackages: []string{},
			Odoobase:         []string{},
		},
		{
			Name:         "18.0",
			Version:      "18.0",
			Image:        "ubuntu/24.04",
			InstanceName: "odoo-18-0",
			Repos:        []string{"odoo", "enterprise", "design-themes", "industry"},
			BaselinePackages: []string{
				"apt-transport-https", "apt-utils", "bzip2", "ca-certificates", "curl",
				"dirmngr", "git", "gnupg", "inetutils-ping", "libgnutls-dane0", "libgts-bin",
				"libpaper-utils", "locales", "lsb-release", "nodejs", "npm", "odaserver",
				"openssh-server", "postgresql-common", "python3", "python3-full",
				"shared-mime-info", "sudo", "unzip", "vim", "wget", "xz-utils", "zip", "zstd",
			},
			Odoobase: []string{
				"fonts-liberation", "fonts-noto", "fonts-noto-cjk", "fonts-noto-mono",
				"geoip-database", "gsfonts", "python3-asn1crypto", "python3-babel",
				"python3-cbor2", "python3-chardet", "python3-cryptography", "python3-cups",
				"python3-dateutil", "python3-decorator", "python3-docutils",
				"python3-feedparser", "python3-freezegun", "python3-geoip2",
				"python3-gevent", "python3-googleapi", "python3-greenlet",
				"python3-html2text", "python3-idna", "python3-jinja2", "python3-ldap",
				"python3-libsass", "python3-lxml", "python3-lxml-html-clean",
				"python3-markupsafe", "python3-num2words", "python3-odf", "python3-ofxparse",
				"python3-olefile", "python3-openpyxl", "python3-openssl", "python3-paramiko",
				"python3-passlib", "python3-pdfminer", "python3-phonenumbers", "python3-pil",
				"python3-pip", "python3-polib", "python3-psutil", "python3-psycopg2",
				"python3-pydot", "python3-pylibdmtx", "python3-pyparsing", "python3-pypdf2",
				"python3-qrcode", "python3-renderpm", "python3-reportlab",
				"python3-rl-renderpm", "python3-reportlab-accel", "python3-requests",
				"python3-rjsmin", "python3-serial", "python3-setuptools", "python3-stdnum",
				"python3-tz", "python3-urllib3", "python3-usb", "python3-vobject",
				"python3-werkzeug", "python3-xlrd", "python3-xlsxwriter", "python3-xlwt",
				"python3-zeep",
			},
		},
	}
}

func GetVersion(version string) *Branch {
	for _, branch := range GetBranches() {
		if branch.Version == version {
			return branch
		}
	}
	return nil
}

func GetBranchLatest() *Branch {
	branchVersions := []string{}
	for _, branch := range GetBranches() {
		branchVersions = append(branchVersions, branch.Version)
	}
	branchVersionMax := slices.Max(branchVersions)
	return GetVersion(branchVersionMax)
}

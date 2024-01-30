package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ppreeper/oda"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "odanode",
		Usage:                "Odoo Node Administration Tool",
		Version:              "0.4.6",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:     "start",
				Usage:    "start odoo server",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					// sudo systemctl start odoo.service
					// return oda.InstanceStart()
					return nil
				},
			},
			{
				Name:     "stop",
				Usage:    "stop odoo server",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					// sudo systemctl stop odoo.service
					// return oda.InstanceStop()
					return nil
				},
			},
			{
				Name:     "restart",
				Usage:    "restart odoo server",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					// sudo systemctl restart odoo.service
					// return oda.InstanceRestart()
					return nil
				},
			},
			{
				Name:     "logs",
				Usage:    "follow the logs",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					// bash -c 'tail -f /var/log/syslog | grep "Odoo Server"'
					// return oda.InstanceLogs()
					return nil
				},
			},
			{
				Name:     "mount",
				Usage:    "mount project and odoo to system",
				Category: "instance",
				Action: func(cCtx *cli.Context) error {
					// BASE=/opt/odoo
					// BSHARE=/share/backups
					// OSHARE=/share/odoo
					// PSHARE=/share/projects
					// COPTS="rsize=8192,wsize=8192,timeo=15"
					// ROPTS="ro,async,noatime"
					// WOPTS="rw,sync,relatime"

					// # stop odoo service
					// sudo systemctl stop odoo.service
					// # mkdir mount directories
					// sudo bash -c "mkdir -p ${BASE}/{addons,conf,data,backups,odoo,enterprise} && chown odoo:odoo ${BASE}"

					// cat <<-_EOF_ | tee /tmp/odoo_fstab > /dev/null
					// #BEGINODOO
					// odoofs:${BSHARE} ${BASE}/backups nfs4 ${WOPTS},${COPTS} 0 0

					// odoofs:${OSHARE}/${args[version]}.0/odoo ${BASE}/odoo nfs4 ${ROPTS},${COPTS} 0 0
					// odoofs:${OSHARE}/${args[version]}.0/enterprise ${BASE}/enterprise nfs4 ${ROPTS},${COPTS} 0 0

					// odoofs:${PSHARE}/${args[projectname]}/${args[branch]}/addons ${BASE}/addons nfs4 ${ROPTS},${COPTS} 0 0
					// odoofs:${PSHARE}/${args[projectname]}/${args[branch]}/conf ${BASE}/conf nfs4 ${ROPTS},${COPTS} 0 0
					// odoofs:${PSHARE}/${args[projectname]}/${args[branch]}/data ${BASE}/data nfs4 ${WOPTS},${COPTS} 0 0
					// #ENDODOO
					// _EOF_

					// sudo sed -e '/#BEGINODOO/{:a; N; /\n#ENDODOO$/!ba; r /tmp/odoo_fstab' -e 'd;}' -i /etc/fstab
					// sudo rm -f /tmp/odoo_fstab

					// sudo umount -R -q ${BASE}/{addons,conf,data,backups,odoo,enterprise}
					// sudo mount -a
					// # restart odoo service
					// sudo systemctl stop odoo.service
					// return oda.InstanceLogs()
					return nil
				},
			},
			{
				Name:     "backup",
				Usage:    "Backup database filestore and addons",
				Category: "admin",
				Action: func(ctx *cli.Context) error {
					// BASE=/opt/odoo
					// cd ${BASE}
					// sudo -u odoo python3 -B /usr/local/bin/oda_db.py -b
					return oda.AdminBackup()
				},
			},
			{
				Name:     "restore",
				Usage:    "Restore database and filestore or addons",
				Category: "admin",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "move",
						Value: false,
						Usage: "move server",
					},
				},
				Action: func(ctx *cli.Context) error {
					// inspect_args
					// [[ -z ${args[--remote]} ]] && REMOTE="" || REMOTE="--remote"
					// BASE=/opt/odoo
					// cd ${BASE}
					// for bfile in ${args[file]}
					// do
					//   sudo -u odoo python3 -B /usr/local/bin/oda_db.py ${REMOTE} -r -d "${bfile}"
					// done
					move := ctx.Bool("move")
					return oda.AdminRestoreNode(move)
				},
			},
			{
				Name:     "requirements",
				Usage:    "install requirements to run Odoo (RUN THIS FIRST)",
				Category: "admin",
				Action: func(ctx *cli.Context) error {
					fmt.Println("installing requirements")
					// sudo bash -c "apt-get update -y"
					// sudo apt-get install -y wget git
					// REPO="https://github.com/wkhtmltopdf/packaging"
					// vers=$(git ls-remote --tags ${REPO} | grep "refs/tags.*[0-9]$" | awk '{print $2}' | sed 's/refs\/tags\///' | sort -V | uniq | tail -1)
					// VC=$(grep ^VERSION_CODENAME /etc/os-release | awk -F'=' '{print $2}')
					// UC=$(grep ^UBUNTU_CODENAME /etc/os-release | awk -F'=' '{print $2}')
					// CN=''
					// [ -n "$UC" ] && CN=$UC || CN=$VC
					// FN="wkhtmltox_${vers}.${CN}_amd64.deb"

					// sudo bash -c "groupadd -f -g 1001 odoo"
					// if [ -f $(grep "^odoo:" /etc/passwd) ]; then
					//   sudo bash -c "useradd -ms /usr/sbin/nologin -g 1001 -u 1001 odoo"
					// fi

					// # setup fstab
					// sudo sed -e '/#BEGINODOO/{:a; N; /\n#ENDODOO$/!ba; echo ""' -e 'd;}' -i /etc/fstab
					// echo "#BEGINODOO" | sudo tee -a /etc/fstab
					// echo "#ENDODOO" | sudo tee -a /etc/fstab

					// # setup hosts cloudinit template
					// sudo sed -e '/#BEGINODOO/{:a; N; /\n#ENDODOO$/!ba; echo ""' -e 'd;}' -i /etc/cloud/templates/hosts.debian.tmpl
					// echo "#BEGINODOO" | sudo tee -a /etc/cloud/templates/hosts.debian.tmpl
					// echo "#ENDODOO" | sudo tee -a /etc/cloud/templates/hosts.debian.tmpl

					// cat <<-_EOF_ | tee /tmp/odoo_hosts.debian.tmpl > /dev/null
					// #BEGINODOO
					// ${args[fsip]} odoofs
					// ${args[dbip]} db
					// ${args[loggerip]} logger
					// #ENDODOO
					// _EOF_

					// sudo sed -e '/#BEGINODOO/{:a; N; /\n#ENDODOO$/!ba; r /tmp/odoo_hosts.debian.tmpl' -e 'd;}' -i /etc/cloud/templates/hosts.debian.tmpl
					// sudo rm -f /tmp/odoo_hosts.debian.tmpl

					// # install prerequisites
					// sudo bash -c "apt-get update -y"
					// sudo apt-get install -y nfs-common qemu-guest-agent

					// # PostgreSQL Repo
					// sudo wget -qO /etc/apt/trusted.gpg.d/pgdg.gpg.asc https://www.postgresql.org/media/keys/ACCC4CF8.asc
					// echo "deb http://apt.postgresql.org/pub/repos/apt/ ${CN}-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list
					// sudo bash -c "apt-get update -y"

					// # postgresql
					// sudo apt-get install -y --no-install-recommends postgresql-client-15

					// # install wkhtmltopdf
					// wget -qc ${REPO}/releases/download/${vers}/${FN} -O ${HOME}/wkhtmltox.deb
					// sudo apt-get install -y --no-install-recommends ${HOME}/wkhtmltox.deb
					// rm -rf ${HOME}/wkhtmltox.deb
					// sudo bash -c "apt-get update -y"

					// # install requirements
					// sudo apt-get install -y --no-install-recommends \
					//     bzip2 \
					//     ca-certificates \
					//     curl \
					//     dirmngr \
					//     fonts-liberation \
					//     fonts-noto \
					//     fonts-noto-cjk \
					//     fonts-noto-mono \
					//     geoip-database \
					//     gnupg \
					//     gsfonts \
					//     inetutils-ping \
					//     libgnutls-dane0 \
					//     libgts-bin \
					//     libpaper-utils \
					//     locales \
					//     nodejs \
					//     npm \
					//     python3 \
					//     python3-babel \
					//     python3-chardet \
					//     python3-cryptography \
					//     python3-cups \
					//     python3-dateutil \
					//     python3-decorator \
					//     python3-docutils \
					//     python3-feedparser \
					//     python3-freezegun \
					//     python3-geoip2 \
					//     python3-gevent \
					//     python3-greenlet \
					//     python3-html2text \
					//     python3-idna \
					//     python3-jinja2 \
					//     python3-ldap \
					//     python3-libsass \
					//     python3-lxml \
					//     python3-markupsafe \
					//     python3-num2words \
					//     python3-ofxparse \
					//     python3-olefile \
					//     python3-openssl \
					//     python3-paramiko \
					//     python3-passlib \
					//     python3-pdfminer \
					//     python3-phonenumbers \
					//     python3-pil \
					//     python3-pip \
					//     python3-polib \
					//     python3-psutil \
					//     python3-psycopg2 \
					//     python3-pydot \
					//     python3-pylibdmtx \
					//     python3-pyparsing \
					//     python3-pypdf2 \
					//     python3-pytzdata \
					//     python3-qrcode \
					//     python3-renderpm \
					//     python3-reportlab \
					//     python3-reportlab-accel \
					//     python3-requests \
					//     python3-rjsmin \
					//     python3-serial \
					//     python3-setuptools \
					//     python3-stdnum \
					//     python3-urllib3 \
					//     python3-usb \
					//     python3-vobject \
					//     python3-werkzeug \
					//     python3-xlrd \
					//     python3-xlsxwriter \
					//     python3-xlwt \
					//     python3-zeep \
					//     shared-mime-info \
					//     unzip \
					//     xz-utils \
					//     zip

					// # install geolite databases
					// sudo wget -qO /usr/share/GeoIP/GeoLite2-ASN.mmdb https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb
					// sudo wget -qO /usr/share/GeoIP/GeoLite2-City.mmdb https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb
					// sudo wget -qO /usr/share/GeoIP/GeoLite2-Country.mmdb https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb

					// # install additional python libraries
					// #sudo pip install --upgrade pip
					// sudo pip3 install --break-system-packages ebaysdk google-auth

					// # install additional node libraries
					// sudo npm -g i rtlcss

					// # update system
					// sudo bash -c "apt-get update -y && apt-get dist-upgrade -y && apt-get autoremove -y && apt-get autoclean -y"

					// # mkdir mount directories
					// sudo bash -c "mkdir -p /opt/odoo/{addons,conf,data,backups,odoo,enterprise} && chown odoo:odoo /opt/odoo"

					// return oda.AdminRestore()
					return nil
				},
			},
			{
				Name:     "systemd",
				Usage:    "install systemd startup script",
				Category: "admin",
				Action: func(ctx *cli.Context) error {
					fmt.Println("installing systemd startup script")
					// function systemd_config(){
					// cat <<-_EOF_ | sudo tee /etc/systemd/system/odoo.service > /dev/null
					// [Unit]
					// Description=Odoo
					// After=remote-fs.target

					// [Service]
					// Type=simple
					// SyslogIdentifier=odoo
					// PermissionsStartOnly=true
					// User=odoo
					// Group=odoo
					// ExecStart=/opt/odoo/odoo/odoo-bin -c /opt/odoo/conf/odoo.conf
					// StandardOutput=journal+console

					// [Install]
					// WantedBy=remote-fs.target
					// _EOF_
					// }

					// systemd_config
					// sudo systemctl daemon-reload
					// sudo systemctl enable odoo.service
					// return oda.AdminRestore()
					return nil
				},
			},
			{
				Name:     "logger",
				Usage:    "redirect rsyslog to logger",
				Category: "admin",
				Action: func(ctx *cli.Context) error {
					fmt.Println("redirecting rsyslog to logger")
					// cat <<-_EOF_ | sudo tee /etc/rsyslog.d/99-logger.conf > /dev/null
					// *.*;auth,authpriv.none    @logger:514
					// _EOF_

					// sudo systemctl restart rsyslog.service
					// return oda.AdminRestore()
					return nil
				},
			},
			{
				Name:     "proxy",
				Usage:    "Caddy proxy",
				Category: "proxy",
				Subcommands: []*cli.Command{
					{
						Name:  "start",
						Usage: "proxy start",
						Action: func(cCtx *cli.Context) error {
							return oda.ProxyStart()
						},
					},
					{
						Name:  "stop",
						Usage: "proxy stop",
						Action: func(cCtx *cli.Context) error {
							return oda.ProxyStop()
						},
					},
					{
						Name:  "restart",
						Usage: "proxy restart",
						Action: func(cCtx *cli.Context) error {
							return oda.ProxyRestart()
						},
					},
					{
						Name:  "generate",
						Usage: "proxy generate",
						Action: func(cCtx *cli.Context) error {
							return oda.ProxyGenerate()
						},
					},
				},
			},
			{
				Name:     "admin",
				Usage:    "Admin user management",
				Category: "instance",
				Subcommands: []*cli.Command{
					{
						Name:  "username",
						Usage: "Odoo Admin username",
						Action: func(cCtx *cli.Context) error {
							return oda.AdminUsername()
						},
					},
					{
						Name:  "password",
						Usage: "Odoo Admin password",
						Action: func(cCtx *cli.Context) error {
							return oda.AdminPassword()
						},
					},
				},
			},
			{
				Name:     "init",
				Usage:    "initialize oda setup",
				Category: "admin",
				Action: func(ctx *cli.Context) error {
					return oda.AdminInit()
				},
			},
			{
				Name:     "project",
				Usage:    "Project level commands [CAUTION]",
				Category: "admin",
				Subcommands: []*cli.Command{
					{
						Name:  "reset",
						Usage: "reset project dir and db",
						Action: func(cCtx *cli.Context) error {
							return oda.ProjectReset()
						},
					},
				},
			},
			{
				Name:     "repo",
				Usage:    "Odoo community and enterprise repository management",
				Category: "admin",
				Subcommands: []*cli.Command{
					{
						Name:  "base",
						Usage: "Odoo Source Repository",
						Subcommands: []*cli.Command{
							{
								Name:  "clone",
								Usage: "clone Odoo source repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBaseClone()
								},
							},
							{
								Name:  "update",
								Usage: "update Odoo source repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBaseUpdate()
								},
							},
						},
					},
					{
						Name:  "branch",
						Usage: "Odoo Source Branch",
						Subcommands: []*cli.Command{
							{
								Name:  "clone",
								Usage: "clone Odoo branch repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBranchClone()
								},
							},
							{
								Name:  "update",
								Usage: "update Odoo branch repository",
								Action: func(cCtx *cli.Context) error {
									return oda.RepoBranchUpdate()
								},
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

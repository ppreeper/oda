#!/bin/bash
[ -z $ODOODB ] && export ODOODB=${PWD##*/}
POD=${PWD##*/}
IPV4=$(ip -4 -br a show | grep -v ^lo | grep UP | awk '{print $3}' | awk -F'/' '{print $1}')
BASE=`dirname "${0}"`

function envrc(){
[ -z ${2} ] && export PORT=8069 || export PORT=${2}
cat << EOF | tee .envrc > /dev/null
layout python3
export ODOO_V=${1}.0
export ODOO_PORT=${PORT}
export ODOO_C=${HOME}/workspace/repos/${1}.0/odoo
export ODOO_E=${HOME}/workspace/repos/${1}.0/enterprise
EOF
}

function configfile(){
cat << EOF | tee conf/odoo.conf > /dev/null
[options]
addons_path = ./odoo/addons,./enterprise,./addons
admin_passwd = adminadmin
without_demo = all
csv_internal_sep = ;
data_dir = ./data
db_host = ${IPV4}
db_port = 5432
db_maxconn = 24
db_user = odoo${1}
db_password = odooodoo
db_name = ${ODOODB}
db_template = template0
db_sslmode = disable
list_db = False
workers = 4
max_cron_threads = 2
http_enable = True
http_interface =
http_port = 8069
reportgz = False
server_wide_modules = base,web
logfile = ./odoo.log
log_level = debug
logrotate = True
# log_handler = werkzeug:CRITICAL,odoo.api:DEBUG
# log_db_level = warning
EOF
}

function pipfile(){
cat << EOF | tee Pipfile > /dev/null
[[source]]
url = "https://pypi.org/simple"
verify_ssl = true
name = "pypi"

[packages]
babel = "==2.9.1"
chardet = "==3.0.4"
cryptography = "==2.6.1"
decorator = "==4.4.2"
docutils = "==0.16"
ebaysdk = "==2.1.5"
freezegun = "==0.3.15"
gevent = "==21.8.0"
google-auth = "==2.17.*"
greenlet = "==1.1.*"
idna = "==2.8"
jinja2 = "==2.11.3"
libsass = "==0.18.0"
lxml = "==4.6.5"
markupsafe = "==1.1.0"
num2words = "==0.5.6"
ofxparse = "==0.21"
passlib = "==1.7.3"
pillow = "==9.0.1"
polib = "==1.1.0"
"pdfminer.six" = "*"
psutil = "==5.6.7"
psycopg2-binary = "2.9.5"
pydot = "==1.4.1"
pyopenssl = "==19.0.0"
pypdf2 = "==1.26.0"
pyserial = "==3.4"
python-dateutil = "==2.8.2"
python-stdnum = "==1.13"
pytz = "*"
pyusb = "==1.0.2"
qrcode = "==6.1"
reportlab = "==3.5.59"
requests = "==2.25.1"
urllib3 = "==1.26.5"
vobject = "==0.9.6.1"
werkzeug = "==2.0.3"
xlrd = "==1.2.0"
xlsxwriter = "==1.1.2"
xlwt = "==1.3.*"
zeep = "==3.4.0"
paramiko = "==2.12.0"

[dev-packages]
yapf="*"
pylint="*"
pylint-odoo="*"

[requires]
python_version = "3"
EOF
}

function initproject(){
    case ${1} in
      "15"|"16")
        envrc ${1} ${2}
        direnv allow >/dev/null
        source .envrc 2>/dev/null
        mkdir -p addons conf data backups
        chmod 777 data backups
        ln -s ${ODOO_C} odoo
        ln -s ${ODOO_E} enterprise
        configfile ${1}
        pipfile
        ;;
    *) echo "invalid odoo version";;
  esac
}

function odoo_install(){
    odoo/odoo-bin -c conf/odoo.conf --no-http --stop-after-init -d $1 -i $2
}

function odoo_upgrade(){
    odoo/odoo-bin -c conf/odoo.conf --no-http --stop-after-init -d $1 -u $2
}

function odoo_start(){
    truncate -s 0 odoo.log
    nohup ../${1}/odoo/odoo-bin -c conf/odoo.conf --http-port $ODOO_PORT > /dev/null &
}

function odoo_stop(){
    pkill -f "${1}/odoo/odoo-bin"
}

function gcfg(){
  # Get variable from odoo config
  printf $(grep "${1}" conf/odoo.conf | awk -F'=' '{print $2}')
}

function cli_help(){
  printf "oda - Odoo Administration Tool\n"
  printf "\nProject Admin Commands\n"
  printf "   initproject          Create a new project\n"
  printf "   destroy              Fully Destroy the project and its files [CAUTION]\n"
  printf "   reset                Drop database and filestore [CAUTION]\n"
  printf "   backup               Backup database and filestore\n"
  printf "   restore <dump_file>  Restore database and filestore [CAUTION]\n"
  printf "\nDatabase Application Commands\n"
  printf "   init               Initialize the database\n"
  printf "   install <modules>  Install module(s) (comma seperated list)\n"
  printf "   upgrade <modules>  Upgrade module(s) (comma seperated list)\n"
  printf "\nDatabase Admin\n"
  printf "   start          Start the instance\n"
  printf "   stop           Stop the instance\n"
  printf "   restart        Restart the instance\n"
  printf "   logs           Follow the logs\n"
  printf "   bin <command>  Run an odoo-bin command\n"
  printf "   psql           Access the raw database\n"
}

case ${1} in
  "init") odoo_install $ODOODB base,l10n_ca && odoo_stop $POD && sleep 2 && odoo_start $POD ;;
  "install") odoo_install $ODOODB ${2} && odoo_stop $POD && sleep 2 && odoo_start $POD ;;
  "upgrade") odoo_upgrade $ODOODB ${2} && odoo_stop $POD && sleep 2 && odoo_start $POD ;;
  "backup" ) python3 -B ${BASE}/oda_db.py -b  ;;
  "restore" ) shift && python3 -B ${BASE}/oda_db.py -r -d "${1}" ;;
  "logs") tail -f odoo.log ;;
  "bin") shift && odoo/odoo-bin $@ ;;
  "start") odoo_start $POD ;;
  "stop") odoo_stop $POD ;;
  "restart") odoo_stop $POD && sleep 2 && odoo_start $POD ;;
  "psql")
    PGPASSWORD=$(gcfg db_pass) psql -U $(gcfg db_user) -h $(gcfg db_host) -p $(gcfg db_port) $(gcfg db_name)
    ;;
  "reset")
    read -r -p "Are you sure? [y/N] " response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        odoo_stop $POD
        rm -rf data/* > /dev/null
        PGPASSWORD=$(gcfg db_pass) dropdb -U $(gcfg db_user) -h $(gcfg db_host) -p $(gcfg db_port) -f $(gcfg db_name)
    fi
    ;;
  "destroy")
    read -r -p "Are you sure you want to destroy everything? [YES/N] " response
    if [[ "$response" =~ ^(YES)$ ]]; then
      read -r -p "Are you **really** sure you want to destroy everything? [YES/N] " response
      if [[ "$response" =~ ^(YES)$ ]]; then
        echo "Destroying project"
        odoo_stop $POD
        sudo rm -rf .direnv/ addons/ backups/ conf/ data/ .envrc Pipfile enterprise odoo
        PGPASSWORD=$(gcfg db_pass) dropdb -U $(gcfg db_user) -h $(gcfg db_host) -p $(gcfg db_port) -f  $(gcfg db_name) >/dev/null
        echo "Project has been destroyed"
      fi
    fi
    ;;
  "initproject") initproject $2 ;;
  *) cli_help ;;
esac

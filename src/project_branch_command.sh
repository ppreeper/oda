function envrc(){
[ -z ${2} ] && export PORT=8069 || export PORT=${2}
cat <<-_EOF_ | tee .envrc > /dev/null
layout python3
export ODOO_V=${1}.0
export ODOO_PORT=${PORT}
export ODOO_C=${ODOOBASE}/${1}.0/odoo
export ODOO_E=${ODOOBASE}/${1}.0/enterprise
_EOF_
}

function configfile(){
cat <<-_EOF_ | tee conf/odoo.conf > /dev/null
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
db_name = ${2}
db_template = template0
db_sslmode = disable
list_db = False
workers = 0
#max_cron_threads = 2
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
_EOF_
}

function pipfile(){
cat <<-_EOF_ | tee Pipfile > /dev/null
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
black="*"
yapf="*"
pylint="*"
pylint-odoo="*"

[requires]
python_version = "3"
_EOF_
}

PDIR=${HOME}/workspace/odoo/${args[projectname]}_${args[branch]}
mkdir -p ${PDIR}
cd ${PDIR}
envrc ${args[version]} ${args[oport]}
direnv allow >/dev/null
mkdir -p conf data backups
git clone ${args[url]} -b ${args[branch]} addons
ODOO_C=$(grep ODOO_C .envrc | awk '{print $2}' | awk -F'=' '{print $2}')
ODOO_E=$(grep ODOO_E .envrc | awk '{print $2}' | awk -F'=' '{print $2}')
if [ -L "odoo" ]; then rm -f odoo ; fi
ln -f -s ${ODOO_C} odoo
if [ -L "enterprise" ]; then rm -f enterprise ; fi
ln -f -s ${ODOO_E} enterprise
configfile ${args[version]} ${args[projectname]}_${args[branch]}
pipfile
printf "To install python dev dependencies run:\npipenv install --dev\n\n"
function configfile(){
cat <<-_EOF_ | sudo -u odoo tee /share/projects/${args[projectname]}/${args[branch]}/conf/odoo.conf > /dev/null
[options]
addons_path = /opt/odoo/odoo/addons,/opt/odoo/enterprise,/opt/odoo/addons
data_dir = /opt/odoo/data
admin_passwd = adminadmin
without_demo = all
csv_internal_sep = ;
db_host = db
db_port = 5432
db_maxconn = 24
db_user = odoo${1}
db_password = odooodoo
db_name = ${args[projectname]}_${args[branch]}
db_template = template0
db_sslmode = disable
list_db = False
workers = 4
max_cron_threads = 2
proxy = True
proxy_mode = True
http_enable = True
http_interface =
http_port = 8069
reportgz = False
syslog = True
log_level = debug
# log_handler = werkzeug:CRITICAL,odoo.api:DEBUG
# log_db_level = warning
_EOF_
}

sudo -u odoo mkdir -p /share/projects/${args[projectname]}/${args[branch]}/{data,conf}
sudo -u odoo git clone ${args[projecturl]} -b ${args[branch]} /share/projects/${args[projectname]}/${args[branch]}/addons
configfile ${args[version]}
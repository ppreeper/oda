odoo/odoo-bin -c conf/odoo.conf --no-http --stop-after-init -d ${args[--name]} -i ${args[modules]}
pkill -f "${POD}/odoo/odoo-bin"
sleep 2
truncate -s 0 odoo.log
nohup ../${POD}/odoo/odoo-bin -c conf/odoo.conf --http-port ${ODOO_PORT} > /dev/null 2>&1 & disown

truncate -s 0 odoo.log
nohup ../${POD}/odoo/odoo-bin -c conf/odoo.conf --http-port ${ODOO_PORT} > /dev/null 2>&1 & disown
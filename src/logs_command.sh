LOG=$(grep logfile conf/odoo.conf | awk -F'=' '{print $2}' | tr -d '[:space:]')
tail -f ${LOG}
if [ -z $ODOO_PORT ]; then
  if [[ -f ".envrc" ]]; then
    export ODOO_PORT=$(grep ODOO_PORT .envrc | awk '{print $2}' | awk -F'=' '{print $2}')
  else
    export ODOO_PORT=$(grep http_port conf/odoo.conf | awk -F'=' '{print $2}' | tr -d '[:space:]')
  fi
fi
truncate -s 0 odoo.log
nohup ../${POD}/odoo/odoo-bin -c conf/odoo.conf --http-port ${ODOO_PORT} > /dev/null 2>&1 & disown
if [[ -f "./conf/odoo.conf" ]]; then
  LOG=$(grep logfile ./conf/odoo.conf | awk -F'=' '{print $2}' | tr -d '[:space:]')
  tail -f ${LOG}
else
  echo "not in a project directory"
fi
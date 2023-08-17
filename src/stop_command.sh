if [[ -f "./conf/odoo.conf" ]]; then
  pkill -f "${POD}/odoo/odoo-bin"
else
  echo "not in a project directory"
fi
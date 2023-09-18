if [[ -f "./conf/odoo.conf" ]]; then
  ../${POD}/odoo/odoo-bin scaffold ${args[module]} ../${POD}/addons/.
else
  echo "not in a project directory"
fi
if [[ -f "./conf/odoo.conf" ]]; then
  ../${POD}/odoo/odoo-bin scaffold ${args[module]} ../${POD}/addons-custom/.
else
  echo "not in a project directory"
fi
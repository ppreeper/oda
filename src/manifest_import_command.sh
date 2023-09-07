if [[ -f "./conf/odoo.conf" ]]; then
  BASE=`dirname "${0}"`
  python3 -B ${BASE}/oda_db.py -i
else
  echo "not in a project directory"
fi
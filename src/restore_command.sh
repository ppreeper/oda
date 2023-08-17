if [[ -f "./conf/odoo.conf" ]]; then
  BASE=`dirname "${0}"`
  for bfile in ${args[file]}
  do
    python3 -B ${BASE}/oda_db.py -r -d "${bfile}"
  done
else
  echo "not in a project directory"
fi
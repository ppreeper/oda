[[ -z ${args[--remote]} ]] && REMOTE="" || REMOTE="--remote"
BASE=/opt/odoo
cd ${BASE}
for bfile in ${args[file]}
do
  sudo -u odoo python3 -B /usr/local/bin/oda_db.py ${REMOTE} -r -d "${bfile}"
done
echo "# this file is located in 'src/remote_restore_command.sh'"
echo "# code for 'oda remote restore' goes here"
echo "# you can edit it freely and regenerate (it will not be overwritten)"
inspect_args

echo ${args[node]}
echo ${args[remote]}
echo ${args[file]}

# DUMP_FILE="backups/2023_08_14_04_13_58_Dec30data.tar.zst"
# PROJECT=$(ssh ${args[node]} -- cat /etc/fstab | grep addons | awk '{print $1}' | awk -F':' '{print $2}' | sed 's,/addons$,,')

# restore filestore on odoofs
# DUMP_FILE="backups/2023_08_14_04_13_58_Dec30data.tar.zst"
# DB="quest15issue_285"
# echo sudo -u odoo rm -rf ${PROJECT}/data/*
# echo sudo -u odoo mkdir -p ${PROJECT}/data/filestore/${DB}
# echo sudo -u odoo tar -axf /share/${DUMP_FILE} -C ${PROJECT}/data/filestore/${DB} --strip-components=2 ./filestore

# echo ssh ${args[remote]} sudo -u odoo rm -rf ${PROJECT}/data/*


# if [[ -f "./conf/odoo.conf" ]]; then
#   BASE=`dirname "${0}"`
#   for bfile in ${args[file]}
#   do
#     if [[ ${args[--remote]} == 1 ]]; then
#       python3 -B ${BASE}/oda_db.py --remote -r -d "${bfile}"
#     else
#       python3 -B ${BASE}/oda_db.py -r -d "${bfile}"
#     fi
#   done
# else
#   echo "not in a project directory"
# fi
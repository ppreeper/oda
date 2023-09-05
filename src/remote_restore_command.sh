echo "# this file is located in 'src/remote_restore_command.sh'"
echo "# code for 'oda remote restore' goes here"
echo "# you can edit it freely and regenerate (it will not be overwritten)"
inspect_args

REMOTE_FS=odoofs
echo ${args[node]}
echo ${args[remote]}
echo ${args[file]}


## Remote FS
PROJECT=$(ssh ${args[node]} -- cat /etc/fstab | grep addons | awk '{print $1}' | awk -F':' '{print $2}' | sed 's,/addons$,,')
DB=$(ssh ${args[node]} -- grep db_name /opt/odoo/conf/odoo.conf | sed 's/ //g' | awk -F'=' '{print $2}')
# ssh ${REMOTE_FS} -- sudo -u odoo rm -rf ${PROJECT}/data/*
# ssh ${REMOTE_FS} -- sudo -u odoo mkdir -p ${PROJECT}/data/filestore/${DB}
# ssh ${REMOTE_FS} -- sudo -u odoo tar -axf /share/${args[file]} -C ${PROJECT}/data/filestore/${DB} --strip-components=2 ./filestore


## Node
# stop service
ssh ${args[node]} -- sudo systemctl stop odoo.service

# restore db on node
# ssh ${args[node]} -- 'cd /opt/odoo && pwd && sudo -u odoo python3 -B /usr/local/bin/oda_db.py --remote -r -d ${args[file]}'

ssh ${args[node]} oda node restore --remote ${args[file]}

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
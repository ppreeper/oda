## Remote FS
REMOTE_FS=odoofs
PROJECT=$(ssh ${args[node]} -- cat /etc/fstab | grep addons | awk '{print $1}' | awk -F':' '{print $2}' | sed 's,/addons$,,')
DB=$(ssh ${args[node]} -- grep db_name /opt/odoo/conf/odoo.conf | sed 's/ //g' | awk -F'=' '{print $2}')
ssh ${REMOTE_FS} -- sudo -u odoo rm -rf ${PROJECT}/data/*
ssh ${REMOTE_FS} -- sudo -u odoo mkdir -p ${PROJECT}/data/filestore/${DB}
ssh ${REMOTE_FS} -- sudo -u odoo tar -axf /share/${args[file]} -C ${PROJECT}/data/filestore/${DB} --strip-components=2 ./filestore

## Node
# stop service
ssh ${args[node]} -- sudo systemctl stop odoo.service

# restore database via node
ssh ${args[node]} -- "cd /opt/odoo && sudo -u odoo python3 -B /usr/local/bin/oda_db.py --remote -r -d ${args[file]} "

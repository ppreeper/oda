mkdir -p ${ODOOBASE}/${args[version]}
rsync -at --inplace --delete ${ODOOBASE}/odoo/ ${ODOOBASE}/${args[version]}/odoo/
rsync -at --inplace --delete ${ODOOBASE}/enterprise/ ${ODOOBASE}/${args[version]}/enterprise/

cd ${ODOOBASE}/${args[version]}/odoo/
git fetch origin && git checkout ${args[version]} && git pull

cd ${ODOOBASE}/${args[version]}/enterprise/
git fetch origin && git checkout ${args[version]} && git pull
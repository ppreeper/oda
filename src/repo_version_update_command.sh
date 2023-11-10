cd ${ODOOBASE}/${args[version]}/odoo/
git fetch origin && git checkout ${args[version]} && git pull

cd ${ODOOBASE}/${args[version]}/enterprise/
git fetch origin && git checkout ${args[version]} && git pull
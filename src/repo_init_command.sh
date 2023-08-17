[ -z ${args[gitbranch]} ] && export BRANCH=${args[version]}.0 || export BRANCH=master

git clone https://github.com/odoo/odoo -b ${BRANCH} ${ODOOBASE}/${args[version]}.0/odoo
git clone https://github.com/odoo/enterprise -b ${BRANCH} ${ODOOBASE}/${args[version]}.0/enterprise
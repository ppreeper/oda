cd ${ODOOBASE}/odoo/
git fetch origin && git checkout $(git branch -a | grep HEAD | awk '{print $3}' | awk -F'/' '{print $2}') && git pull

cd ${ODOOBASE}/enterprise/
git fetch origin && git checkout $(git branch -a | grep HEAD | awk '{print $3}' | awk -F'/' '{print $2}') && git pull

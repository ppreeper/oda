read -r -p "Are you sure you want to reset the database? [YES/N] " response
if [[ "$response" =~ ^(YES)$ ]]; then
  read -r -p "Are you **really** sure you want to reset the database? [YES/N] " response
  if [[ "$response" =~ ^(YES)$ ]]; then
    echo "Resetting project"
    sudo systemctl stop odoo.service
    sudo -u odoo rm -rf /opt/odoo/data/* > /dev/null
    PGPASSWORD=${args[--pass]} dropdb -U ${args[--user]} -h ${args[--host]} -p ${args[--port]} -w -f ${args[--name]} >/dev/null
    echo "Project reset"
  fi
fi
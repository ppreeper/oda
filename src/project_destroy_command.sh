read -r -p "Are you sure you want to destroy everything? [YES/N] " response
if [[ "$response" =~ ^(YES)$ ]]; then
  read -r -p "Are you **really** sure you want to destroy everything? [YES/N] " response
  if [[ "$response" =~ ^(YES)$ ]]; then
    echo "Destroying project"
    trap "pkill -f ${POD}/odoo/odoo-bin" SIGINT
    rm -rf .direnv/ .envrc *
    PGPASSWORD=${args[--pass]} dropdb -U ${args[--user]} -h ${args[--host]} -p ${args[--port]} -w -f ${args[--name]} >/dev/null
    echo "Project has been destroyed"
  fi
fi
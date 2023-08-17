if [[ -f "./conf/odoo.conf" ]]; then
  psql postgres://${args[--user]}:${args[--pass]}@${args[--host]}:${args[--port]}/${args[--name]}
else
  echo "not in a project directory"
fi
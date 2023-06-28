if [ -z "${args[admin_name]}" ]; then
  admin_user=$(psql postgres://${args[--db_user]}:${args[--db_pass]}@${args[--host]}:${args[--port]}/${args[--db_name]} -t -c "select login from res_users where id=2;")
  echo "Odoo Admin username: $(echo $admin_user || awk '{print $1}')"
else
  read -r -p "Are you sure you want to change the admin username to: ${args[admin_name]} [YES/N] " response
  if [[ "$response" =~ ^(YES)$ ]]; then
  echo "changing username to: ${args[admin_name]}"
  admin_user=$(psql postgres://${args[--db_user]}:${args[--db_pass]}@${args[--host]}:${args[--port]}/${args[--db_name]} -t -c "update res_users set login='${args[admin_name]}' where id=2;")
  fi
fi
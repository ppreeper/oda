if [[ -f "./conf/odoo.conf" ]]; then
  if [ -z $ODOO_PORT ]; then
    if [[ -f ".envrc" ]]; then
      export ODOO_PORT=$(grep ODOO_PORT .envrc | awk '{print $2}' | awk -F'=' '{print $2}')
    else
      export ODOO_PORT=$(grep http_port conf/odoo.conf | awk -F'=' '{print $2}' | tr -d '[:space:]')
    fi
  fi

  EXEC="odooquery -host localhost -port ${ODOO_PORT} -d ${args[--db_name]} -U ${args[-U]} -P ${args[-P]} -model ${args[model]}"
  [ -z ${args[--filter]} ] || EXEC="${EXEC} -filter \"${args[--filter]}\""
  [ -z ${args[--fields]} ] || EXEC="${EXEC} -fields ${args[--fields]}"
  [ -z ${args[--limit]} ] || EXEC="${EXEC} -limit ${args[--limit]}"
  [ -z ${args[--offset]} ] || EXEC="${EXEC} -offset ${args[--offset]}"
  [ -z ${args[--count]} ] || EXEC="${EXEC} -count"
  ${EXEC}
else
  echo "not in a project directory"
fi
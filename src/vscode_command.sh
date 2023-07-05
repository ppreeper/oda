if [ -z $ODOO_PORT ]; then
  if [[ -f ".envrc" ]]; then
    export ODOO_PORT=$(grep ODOO_PORT .envrc | awk '{print $2}' | awk -F'=' '{print $2}')
  else
    export ODOO_PORT=$(grep http_port conf/odoo.conf | awk -F'=' '{print $2}' | tr -d '[:space:]')
  fi
fi

function launch_json(){
[ -z ${2} ] && export PORT=8069 || export PORT=${2}
cat <<-_EOF_ | tee .vscode/launch.json > /dev/null
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch",
      "type": "python",
      "request": "launch",
      "stopOnEntry": false,
      "python": "\${command:python.interpreterPath}",
      "program": "\${workspaceRoot}/odoo/odoo-bin",
      "args": ["-c", "\${workspaceRoot}/conf/odoo.conf","-p","$ODOO_PORT"],
      "cwd": "\${workspaceRoot}",
      "env": {},
      "envFile": "\${workspaceFolder}/.env",
      "console": "integratedTerminal"
    }
  ]
}
_EOF_
}

function settings_json(){
[ -z ${2} ] && export PORT=8069 || export PORT=${2}
cat <<-_EOF_ | tee .vscode/settings.json > /dev/null
{
  "python.analysis.extraPaths": ["odoo", "enterprise"],
  "python.linting.pylintEnabled": true,
  "python.linting.enabled": true,
  "python.terminal.executeInFileDir": true,
  "python.formatting.provider": "yapf",
  "python.formatting.yapfArgs": [
    "--style",
    "{allow_split_before_dict_value=False,force_multiline_dict=True,split_before_closing_bracket=True,join_multiple_lines=False}"
  ]
}
_EOF_
}

mkdir -p .vscode
settings_json
launch_json
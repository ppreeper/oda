cat <<-_EOF_ | tee pyrightconfig.json > /dev/null
{
  "venvPath": ".",
  "venv": ".direnv",
  "executionEnvironments": [
    {
      "root": ".",
      "extraPaths": ["addons","odoo","odoo/odoo", "enterprise"]
    }
  ]
}
_EOF_

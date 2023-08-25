cat <<-_EOF_ | tee pyrightconfig.json > /dev/null
{
  "executionEnvironments": [
    {
      "root": "addons",
      "extraPaths": ["odoo", "enterprise"]
    }
  ]
}
_EOF_

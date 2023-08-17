function systemd_config(){
cat <<-_EOF_ | sudo tee /etc/systemd/system/odoo.service > /dev/null
[Unit]
Description=Odoo
After=remote-fs.target

[Service]
Type=simple
SyslogIdentifier=odoo
PermissionsStartOnly=true
User=odoo
Group=odoo
ExecStart=/opt/odoo/odoo/odoo-bin -c /opt/odoo/conf/odoo.conf
StandardOutput=journal+console

[Install]
WantedBy=remote-fs.target
_EOF_
}

systemd_config
sudo systemctl daemon-reload
sudo systemctl enable odoo.service
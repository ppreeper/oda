function systemd_config(){
cat <<-_EOF_ | tee /etc/systemd/system/odoo.service > /dev/null
[Unit]
Description=Odoo
Requires=postgresql.service
After=network.target postgresql.service

[Service]
Type=simple
SyslogIdentifier=odoo
PermissionsStartOnly=true
User=odoo
Group=odoo
ExecStart=/opt/odoo/odoo/odoo-bin -c /opt/odoo/conf/odoo.conf
StandardOutput=journal+console

[Install]
WantedBy=multi-user.target
_EOF_
}

systemd_config
sudo systemctl daemon-reload
sudo systemctl enable odoo.service
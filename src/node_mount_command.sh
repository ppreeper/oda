BASE=/opt/odoo
BSHARE=/share/backups
OSHARE=/share/odoo
PSHARE=/share/projects
COPTS="rsize=8192,wsize=8192,timeo=15"
ROPTS="ro,async,noatime"
WOPTS="rw,sync,relatime"

sudo systemctl stop odoo.service
# mkdir mount directories
sudo bash -c "mkdir -p ${BASE}/{addons,conf,data,backups,odoo,enterprise} && chown odoo:odoo ${BASE}"

cat <<-_EOF_ | tee /tmp/odoo_fstab > /dev/null
#BEGINODOO
odoofs:${BSHARE} ${BASE}/backups nfs4 ${WOPTS},${COPTS} 0 0

odoofs:${OSHARE}/${args[version]}.0/odoo ${BASE}/odoo nfs4 ${ROPTS},${COPTS} 0 0
odoofs:${OSHARE}/${args[version]}.0/enterprise ${BASE}/enterprise nfs4 ${ROPTS},${COPTS} 0 0

odoofs:${PSHARE}/${args[projectname]}/${args[branch]}/addons ${BASE}/addons nfs4 ${ROPTS},${COPTS} 0 0
odoofs:${PSHARE}/${args[projectname]}/${args[branch]}/conf ${BASE}/conf nfs4 ${ROPTS},${COPTS} 0 0
odoofs:${PSHARE}/${args[projectname]}/${args[branch]}/data ${BASE}/data nfs4 ${WOPTS},${COPTS} 0 0
#ENDODOO
_EOF_

sudo sed -e '/#BEGINODOO/{:a; N; /\n#ENDODOO$/!ba; r /tmp/odoo_fstab' -e 'd;}' -i /etc/fstab
sudo rm -f /tmp/odoo_fstab

sudo umount ${BASE}/{addons,conf,data,backups,odoo,enterprise}
sudo mount -a

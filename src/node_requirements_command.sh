REPO="https://github.com/wkhtmltopdf/packaging"
vers=$(git ls-remote --tags ${REPO} | grep "refs/tags.*[0-9]$" | awk '{print $2}' | sed 's/refs\/tags\///' | sort -V | uniq | tail -1)
VC=$(grep ^VERSION_CODENAME /etc/os-release | awk -F'=' '{print $2}')
UC=$(grep ^UBUNTU_CODENAME /etc/os-release | awk -F'=' '{print $2}')
CN=''
[ -n "$UC" ] && CN=$UC || CN=$VC
FN="wkhtmltox_${vers}.${CN}_amd64.deb"

sudo bash -c "groupadd -f -g 1001 odoo"
if [ -f $(grep "^odoo:" /etc/passwd) ]; then
  sudo bash -c "useradd -ms /usr/sbin/nologin -g 1001 -u 1001 odoo"
fi

# setup fstab
sudo sed -e '/#BEGINODOO/{:a; N; /\n#ENDODOO$/!ba; echo ""' -e 'd;}' -i /etc/fstab
echo "#BEGINODOO" | sudo tee -a /etc/fstab
echo "#ENDODOO" | sudo tee -a /etc/fstab

# setup hosts cloudinit template
sudo sed -e '/#BEGINODOO/{:a; N; /\n#ENDODOO$/!ba; echo ""' -e 'd;}' -i /etc/cloud/templates/hosts.debian.tmpl
echo "#BEGINODOO" | sudo tee -a /etc/cloud/templates/hosts.debian.tmpl
echo "#ENDODOO" | sudo tee -a /etc/cloud/templates/hosts.debian.tmpl

cat <<-_EOF_ | tee /tmp/odoo_hosts.debian.tmpl > /dev/null
#BEGINODOO
${args[fsip]} odoofs
${args[dbip]} db
${args[loggerip]} logger
#ENDODOO
_EOF_

sudo sed -e '/#BEGINODOO/{:a; N; /\n#ENDODOO$/!ba; r /tmp/odoo_hosts.debian.tmpl' -e 'd;}' -i /etc/cloud/templates/hosts.debian.tmpl
sudo rm -f /tmp/odoo_hosts.debian.tmpl

# install prerequisites
sudo bash -c "apt-get update -y"
sudo apt-get install -y wget nfs-common qemu-guest-agent

# PostgreSQL Repo
sudo wget -qO /etc/apt/trusted.gpg.d/pgdg.gpg.asc https://www.postgresql.org/media/keys/ACCC4CF8.asc
echo "deb http://apt.postgresql.org/pub/repos/apt/ ${CN}-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list
sudo bash -c "apt-get update -y"

# postgresql
sudo apt-get install -y --no-install-recommends postgresql-client-15

# install wkhtmltopdf
wget -qc ${REPO}/releases/download/${vers}/${FN} -O ${HOME}/wkhtmltox.deb
sudo apt-get install -y --no-install-recommends ${HOME}/wkhtmltox.deb
rm -rf ${HOME}/wkhtmltox.deb
sudo bash -c "apt-get update -y"

# install requirements
sudo apt-get install -y --no-install-recommends \
    bzip2 \
    ca-certificates \
    curl \
    dirmngr \
    fonts-liberation \
    fonts-noto \
    fonts-noto-cjk \
    fonts-noto-mono \
    geoip-database \
    gnupg \
    gsfonts \
    inetutils-ping \
    libgnutls-dane0 \
    libgts-bin \
    libpaper-utils \
    locales \
    nodejs \
    npm \
    python3 \
    python3-babel \
    python3-chardet \
    python3-cryptography \
    python3-cups \
    python3-dateutil \
    python3-decorator \
    python3-docutils \
    python3-feedparser \
    python3-freezegun \
    python3-geoip2 \
    python3-gevent \
    python3-greenlet \
    python3-html2text \
    python3-idna \
    python3-jinja2 \
    python3-ldap \
    python3-libsass \
    python3-lxml \
    python3-markupsafe \
    python3-num2words \
    python3-ofxparse \
    python3-olefile \
    python3-openssl \
    python3-paramiko \
    python3-passlib \
    python3-pdfminer \
    python3-phonenumbers \
    python3-pil \
    python3-pip \
    python3-polib \
    python3-psutil \
    python3-psycopg2 \
    python3-pydot \
    python3-pylibdmtx \
    python3-pyparsing \
    python3-pypdf2 \
    python3-pytzdata \
    python3-qrcode \
    python3-renderpm \
    python3-reportlab \
    python3-reportlab-accel \
    python3-requests \
    python3-rjsmin \
    python3-serial \
    python3-setuptools \
    python3-stdnum \
    python3-urllib3 \
    python3-usb \
    python3-vobject \
    python3-werkzeug \
    python3-xlrd \
    python3-xlsxwriter \
    python3-xlwt \
    python3-zeep \
    shared-mime-info \
    unzip \
    xz-utils \
    zip

# install geolite databases
sudo wget -qO /usr/share/GeoIP/GeoLite2-ASN.mmdb https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb
sudo wget -qO /usr/share/GeoIP/GeoLite2-City.mmdb https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb
sudo wget -qO /usr/share/GeoIP/GeoLite2-Country.mmdb https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb

# install additional python libraries
#sudo pip install --upgrade pip
sudo pip3 install --break-system-packages ebaysdk google-auth

# install additional node libraries
sudo npm -g i rtlcss

# update system
sudo bash -c "apt-get update -y && apt-get dist-upgrade -y && apt-get autoremove -y && apt-get autoclean -y"

# mkdir mount directories
sudo bash -c "mkdir -p /opt/odoo/{addons,conf,data,backups,odoo,enterprise} && chown odoo:odoo /opt/odoo"

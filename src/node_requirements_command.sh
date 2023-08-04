# PostgreSQL Repo
sudo wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O /etc/apt/trusted.gpg.d/pgdg.gpg.asc
echo 'deb http://apt.postgresql.org/pub/repos/apt/ jammy-pgdg main' | sudo tee /etc/apt/sources.list.d/pgdg.list

# update system
sudo bash -c "apt-get update -y && apt-get dist-upgrade -y && apt-get autoremove -y && apt-get autoclean -y"

# install python and node
sudo apt-get install -y --no-install-recommends python3 python3-pip python3-setuptools nodejs npm
# install requirements
sudo apt-get install -y --no-install-recommends \
  bzip2 ca-certificates curl dirmngr fonts-liberation fonts-noto fonts-noto-cjk fonts-noto-mono \
  gnupg gsfonts inetutils-ping libgnutls-dane0 libgts-bin libpaper-utils locales nfs-common \
  postgresql-client-15 shared-mime-info unzip xz-utils zip \
  python3-babel python3-chardet python3-cryptography python3-cups python3-dateutil \
  python3-decorator python3-docutils python3-feedparser python3-freezegun python3-geoip2 \
  python3-gevent python3-greenlet python3-html2text python3-idna python3-jinja2 \
  python3-ldap python3-libsass python3-lxml python3-markupsafe python3-num2words \
  python3-ofxparse python3-olefile python3-openssl python3-paramiko python3-passlib \
  python3-pdfminer python3-phonenumbers python3-pil python3-polib python3-psutil \
  python3-psycopg2 python3-pydot python3-pylibdmtx python3-pyparsing python3-pypdf2 \
  python3-pytzdata python3-qrcode python3-renderpm python3-reportlab python3-reportlab-accel \
  python3-requests python3-rjsmin python3-serial python3-stdnum python3-urllib3 \
  python3-usb python3-vobject python3-werkzeug python3-xlrd python3-xlsxwriter \
  python3-xlwt python3-zeep

# install wkhtmltopdf
wget -q https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltox_0.12.6.1-3.jammy_amd64.deb -O wkhtmltox.deb
sudo apt-get install -y --no-install-recommends ./wkhtmltox.deb \
rm -rf wkhtmltox.deb

# install additional python libraries
sudo pip install --upgrade pip
sudo pip3 install ebaysdk google-auth

# install additional node libraries
sudo npm -g i rtlcss

# mkdir mount directories
sudo bash -c "mkdir -p /opt/odoo/{addons,conf,data,backups,odoo,enterprise} && chown odoo:odoo /opt/odoo"
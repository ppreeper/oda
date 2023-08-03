cat <<-_EOF_ | tee /etc/rsyslog.d/99-logger.conf > /dev/null
*.*;auth,authpriv.none    @logger:514
_EOF_

sudo systemctl restart rsyslog.service
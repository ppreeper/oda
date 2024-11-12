package lib

import (
	"net"
	"os"
	"strings"
)

func GetFQDN() (fqdn, hostname, domain string) {
	hostname, _ = os.Hostname()
	fqdn, _ = net.LookupCNAME(hostname)
	ff := strings.Split(fqdn, ".")
	if len(ff) == 1 {
		return fqdn + ".local", hostname, "local"
	}
	if ff[len(ff)-1] == "" {
		ff = ff[:len(ff)-1]
	}
	domain = strings.Join(ff[1:], ".")
	if domain == "" {
		domain = "local"
	}
	return hostname + "." + domain, hostname, domain
}

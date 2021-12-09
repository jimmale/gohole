package utils

import (
	log "github.com/sirupsen/logrus"
	"net"
	"regexp"
)

func GetLocalIP() string {
	// TODO 127.0.0.53 is probably the only address we shouldn't try binding to.

	localIPRegex := regexp.MustCompile(`192\.168\..*`)
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, a := range addrs {
			ip, _, _ := net.ParseCIDR(a.String())
			if localIPRegex.Match([]byte(ip.To4().String())) {
				log.Printf("Local IP: %s", ip.To4().String())
				return ip.To4().String()
			}
		}
	}
	return ""
}

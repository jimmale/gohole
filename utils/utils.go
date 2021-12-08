package utils

import (
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net"
	"regexp"
)

func Resolve(originalMessage *dns.Msg) *dns.Msg {
	var output *dns.Msg

	// Make a recursive DNS call
	c := new(dns.Client)
	newMessage := new(dns.Msg)
	newMessage.SetQuestion(originalMessage.Question[0].Name, originalMessage.Question[0].Qtype)
	newMessage.RecursionDesired = true
	r, _, err := c.Exchange(newMessage, "1.1.1.1:53")

	if err != nil {
		log.Errorf("Error making recursive DNS Call: %s", err.Error())
	}

	output = r
	r.SetReply(originalMessage)

	return output
}

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

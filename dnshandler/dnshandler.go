package dnshandler

import (
	"bufio"
	"github.com/jimmale/gohole/utils"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

type GoholeHandler struct {
	blocklist map[string]bool
}

func (ghh GoholeHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.MsgHdr.RecursionAvailable = true

	if ghh.DomainIsBlocked(r.Question[0].Name) {
		log.Debugf("%s is blocked", r.Question[0].Name)
		msg.Rcode = dns.RcodeNameError
	} else {
		log.Tracef("%s is not blocked", r.Question[0].Name)
		msg = utils.Resolve(r)
	}
	w.WriteMsg(msg)
	w.Close()
}

func (ghh *GoholeHandler) DomainIsBlocked(domain string) bool {
	_, blocked := ghh.blocklist[domain]
	return blocked
}

func (ghh *GoholeHandler) UpdateBlockList() {
	log.Println("Updating Blocklist")
	if ghh.blocklist == nil {
		log.Trace("blocklist is nil. Instantiating")
		ghh.blocklist = map[string]bool{}
	}

	reg := regexp.MustCompile(`^0\.0\.0\.0 .*`)

	resp, err := http.Get("BlockListURL")
	if err != nil {
		log.Errorf("Error fetching blocklist: %s", err.Error())
		return
	}
	defer resp.Body.Close()
	log.Trace("Downloaded Blocklist. Parsing...")

	sc := bufio.NewScanner(resp.Body)

	for sc.Scan() {
		line := sc.Text()
		if reg.Match([]byte(line)) {
			parts := strings.Split(line, " ")

			// domain names actually have a dot at the end.
			actualFQDN := parts[1] + "."
			ghh.blocklist[actualFQDN] = true
		}
	}
}

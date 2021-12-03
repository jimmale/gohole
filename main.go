package main

import (
	"bufio"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

const BlockListURL = "https://blocklistproject.github.io/Lists/ads.txt"

func main() {

	log.SetFormatter(&log.TextFormatter{
		ForceColors:               false,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             true,
		TimestampFormat:           "2006.01.02 15:04:05.000 Z0700",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    true,
		PadLevelText:              false,
		QuoteEmptyFields:          true,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	})

	log.SetLevel(log.DebugLevel)

	mh := MyHandler{}
	mh.UpdateBlockList()

	log.Trace("Ready.")
	log.Fatalf(dns.ListenAndServe("192.168.1.39:53", "udp4", mh).Error())

}

type MyHandler struct {
	blocklist map[string]bool
}

func (m MyHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.MsgHdr.RecursionAvailable = true

	if m.DomainIsBlocked(r.Question[0].Name) {
		log.Debugf("%s is blocked", r.Question[0].Name)
		msg.Rcode = dns.RcodeNameError
	} else {
		log.Tracef("%s is not blocked", r.Question[0].Name)
		msg = Resolve(r)
	}
	w.WriteMsg(msg)
	w.Close()
}

func Resolve(originalMessage *dns.Msg) *dns.Msg {
	var output *dns.Msg

	// Make a recursive DNS call
	c := new(dns.Client)
	newMessage := new(dns.Msg)
	newMessage.SetQuestion(originalMessage.Question[0].Name, originalMessage.Question[0].Qtype)
	newMessage.RecursionDesired = true
	r, _, err := c.Exchange(newMessage, "8.8.8.8:53")

	if err != nil {
		log.Errorf("Error making recursive DNS Call: %s", err.Error())
	}

	output = r
	r.SetReply(originalMessage)

	return output
}

func (m *MyHandler) DomainIsBlocked(domain string) bool {
	_, blocked := m.blocklist[domain]
	return blocked
}

func (m *MyHandler) UpdateBlockList() {
	if m.blocklist == nil {
		log.Trace("blocklist is nil. Instantiating")
		m.blocklist = map[string]bool{}
	}

	reg := regexp.MustCompile(`^0\.0\.0\.0 .*`)

	resp, err := http.Get(BlockListURL)
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
			m.blocklist[actualFQDN] = true
		}
	}
}

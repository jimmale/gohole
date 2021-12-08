package dnshandler

import (
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

// GoholeHandler is kind of a proxy to the GoholeResolver
type GoholeHandler struct {
	blocklist map[string]bool

	Resolver *GoholeResolver
}

// ServeDNS is the interface we need to satisfy for miekg/dns
func (ghh GoholeHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := ghh.Resolver.Resolve(r)
	w.WriteMsg(&msg)
	w.Close()
}

// GoholeResolver is the cache and blocking mechanism
type GoholeResolver struct {
	// This is the recursive DNS resolver
	UpstreamDNS []string

	IndividualBlockedDomains map[string]interface{} // Domains that the user has explicitly blocked
	IndividualAllowedDomains map[string]interface{} // Domains that the user has explicitly allowed

	Blocklists []string // List of URLs of blocklists
}

func (ghr *GoholeResolver) Resolve(r *dns.Msg) dns.Msg{
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.MsgHdr.RecursionAvailable = true
	if ghr.DomainIsBlocked(r.Question[0].Name) {
		log.Debugf("%s is blocked", r.Question[0].Name)
		msg.Rcode = dns.RcodeNameError
	} else {
		log.Tracef("%s is not blocked", r.Question[0].Name)
		msg = ghr.recursivelyResolve(r)
	}

	// TODO
	return *r
}

func (ghr *GoholeResolver) recursivelyResolve(originalMessage *dns.Msg) *dns.Msg {
	var output *dns.Msg

	// Make a recursive DNS call
	c := new(dns.Client)
	newMessage := new(dns.Msg)
	newMessage.SetQuestion(originalMessage.Question[0].Name, originalMessage.Question[0].Qtype)
	newMessage.RecursionDesired = true
	r, _, err := c.Exchange(newMessage, "1.1.1.1:53") // TODO use configured DNS server

	if err != nil {
		log.Errorf("Error making recursive DNS Call: %s", err.Error())
	}

	output = r
	r.SetReply(originalMessage)

	return output
}


func (ghr *GoholeResolver) DomainIsBlocked(domain string) bool {
	// TODO
		return false
}

//func (ghh *GoholeHandler) UpdateBlockList() {
//	log.Println("Updating Blocklist")
//	if ghh.blocklist == nil {
//		log.Trace("blocklist is nil. Instantiating")
//		ghh.blocklist = map[string]bool{}
//	}
//
//	reg := regexp.MustCompile(`^0\.0\.0\.0 .*`)
//
//	resp, err := http.Get("BlockListURL")
//	if err != nil {
//		log.Errorf("Error fetching blocklist: %s", err.Error())
//		return
//	}
//	defer resp.Body.Close()
//	log.Trace("Downloaded Blocklist. Parsing...")
//
//	sc := bufio.NewScanner(resp.Body)
//
//	for sc.Scan() {
//		line := sc.Text()
//		if reg.Match([]byte(line)) {
//			parts := strings.Split(line, " ")
//
//			// domain names actually have a dot at the end.
//			actualFQDN := parts[1] + "."
//			ghh.blocklist[actualFQDN] = true
//		}
//	}
//}



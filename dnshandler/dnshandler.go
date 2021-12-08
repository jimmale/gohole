package dnshandler

import (
	"fmt"
	"github.com/ReneKroon/ttlcache/v2"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// GoholeHandler is kind of a proxy to the GoholeResolver
type GoholeHandler struct {
	blocklist map[string]bool

	Resolver *GoholeResolver
}

// ServeDNS is the interface we need to satisfy for miekg/dns
func (ghh GoholeHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := ghh.Resolver.Resolve(r)
	w.WriteMsg(msg)
	w.Close()
}

// GoholeResolver is the cache and blocking mechanism
type GoholeResolver struct {
	// This is the recursive DNS resolver
	UpstreamDNS []string

	IndividualBlockedDomains map[string]interface{} // Domains that the user has explicitly blocked
	IndividualAllowedDomains map[string]interface{} // Domains that the user has explicitly allowed

	Blocklists []string // List of URLs of blocklists

	DNSCache *ttlcache.Cache
}

func NewGoholeResolver(c *cli.Context) *GoholeResolver {
	output := GoholeResolver{}

	output.DNSCache = ttlcache.NewCache()

	for _, v := range c.StringSlice("block") {
		output.BlockDomain(v)
	}

	return &output
}

func (ghr *GoholeResolver) Resolve(r *dns.Msg) *dns.Msg {

	cacheKey := CacheKey(r)

	cacheEntry, err := ghr.DNSCache.Get(cacheKey)
	if err != nil {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Rcode = dns.RcodeNameError // NXDomain
		log.Tracef("Couldn't find something in the cache.")
		return msg

	}

	msg := cacheEntry.(*dns.Msg)
	myRcode := msg.Rcode

	msg.SetReply(r) // SetReply resets the
	msg.Rcode = myRcode

	msg.MsgHdr.RecursionAvailable = true
	msg.MsgHdr.Authoritative = true

	// TODO
	return msg
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

func (ghr *GoholeResolver) BlockDomain(domain string) {
	NXDomainMessage := new(dns.Msg)
	NXDomainMessage.Rcode = dns.RcodeNameError // NXDomain error
	NXDomainMessage.RecursionAvailable = true
	NXDomainMessage.Authoritative = true
	domainWithDot := domain + "."

	log.Tracef("Adding %s to DNS Cache as blocked", domain)
	ghr.DNSCache.Set(domainWithDot, NXDomainMessage)
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

func CacheKey(msg *dns.Msg) string {
	var output string
	if len(msg.Answer) != 0 {
		output = fmt.Sprintf("%d:%s", msg.Answer[0].Header().Rrtype, msg.Answer[0].Header().Name)
	} else {
		output = fmt.Sprintf("%d:%s", msg.Question[0].Qtype, msg.Question[0].Name)
	}
	return output
}

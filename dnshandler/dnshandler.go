package dnshandler

import (
	"fmt"
	"github.com/ReneKroon/ttlcache/v2"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"strconv"
	"strings"
	"time"
)

// GoholeHandler is kind of a proxy to the GoholeResolver
type GoholeHandler struct {
	blocklist map[string]bool

	Resolver *GoholeResolver
}

// ServeDNS is the interface we need to satisfy for miekg/dns
func (ghh GoholeHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := ghh.Resolver.Resolve(r)
	_ = w.WriteMsg(msg)
	_ = w.Close()
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

// NewGoholeResolver sets up a new Resolver
func NewGoholeResolver(c *cli.Context) *GoholeResolver {
	output := GoholeResolver{}

	output.DNSCache = ttlcache.NewCache()
	output.DNSCache.SetLoaderFunction(loaderFunction)

	if c != nil {
		// Block all domains on the blocklists
		// TODO

		// Block all individual domains
		for _, v := range c.StringSlice("block") {
			output.BlockDomain(v)
		}

		// Unblock all individual domains
		// TODO

	}
	return &output
}

// Resolve resolves a DNS query and returns a result.
func (ghr *GoholeResolver) Resolve(r *dns.Msg) *dns.Msg {
	cacheKey := CacheKey(r)                     // calculate the cache key (eg "1:example.com." is the key for the A record for example.com)
	cacheEntry, _ := ghr.DNSCache.Get(cacheKey) // fetch from cache; or load from recursive resolver
	msg := cacheEntry.(*dns.Msg)                // cast to the appropriate struct

	myRcode := msg.Rcode // save the Rcode because the dns.Msg.SetReply method sets it to 0 for some reason

	msg.SetReply(r)
	msg.Rcode = myRcode // restore the Rcode
	msg.MsgHdr.RecursionAvailable = true
	msg.MsgHdr.Authoritative = true

	return msg
}

// BlockDomain blocks a domain in the Resolver. All queries for that domain will return NXDomain.
func (ghr *GoholeResolver) BlockDomain(domain string) {
	NXDomainMessage := new(dns.Msg)
	NXDomainMessage.Rcode = dns.RcodeNameError // NXDomain error
	NXDomainMessage.RecursionAvailable = true
	NXDomainMessage.Authoritative = true

	log.Tracef("Adding %s to DNS Cache as blocked", domain)

	for i := 0; i <= 256; i++ {
		cacheKey := fmt.Sprintf("%d:%s.", i, domain)
		_ = ghr.DNSCache.Set(cacheKey, NXDomainMessage)
	}
}

// loaderFunction is a function that is called on a cache miss in order to initiate a recursive resolution and store the
// response in the cache
func loaderFunction(key string) (data interface{}, ttl time.Duration, err error) {
	log.Tracef("Could not find domain in cache. Loading from upstream...")

	parts := strings.Split(key, ":")
	Qtypeint, _ := strconv.Atoi(parts[0])
	Qtypeuint16 := uint16(Qtypeint)
	domain := parts[1]

	upstreamResponse := newRecursiveResolve(Qtypeuint16, domain)

	newTTL := time.Duration(upstreamResponse.Answer[0].Header().Ttl)*time.Second - 1*time.Second

	return upstreamResponse, newTTL, nil
}

func newRecursiveResolve(recordType uint16, domain string) *dns.Msg {
	query := new(dns.Msg)
	query.SetQuestion(domain, recordType)
	query.RecursionDesired = true
	query.Id = dns.Id()

	c := new(dns.Client)                            // TODO could probably put this in the ghr to reuse
	resp, _, err := c.Exchange(query, "1.1.1.1:53") // TODO use configured DNS server
	if err != nil {
		log.Errorf("Error making recursive DNS Call: %s", err.Error())
	}

	return resp
}

//// recursivelyResolve fetches the appropriate record from the upstream DNS server
//func (ghr *GoholeResolver) recursivelyResolve(originalMessage *dns.Msg) *dns.Msg {
//	var output *dns.Msg
//
//	// Make a recursive DNS call
//	c := new(dns.Client) // TODO could probably reuse this
//	newMessage := new(dns.Msg)
//	newMessage.SetQuestion(originalMessage.Question[0].Name, originalMessage.Question[0].Qtype)
//	newMessage.RecursionDesired = true
//	r, _, err := c.Exchange(newMessage, "1.1.1.1:53") // TODO use configured DNS server
//
//	if err != nil {
//		log.Errorf("Error making recursive DNS Call: %s", err.Error())
//	}
//
//	output = r
//	r.SetReply(originalMessage)
//	return output
//}

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

// CacheKey generates a string key on the tuple (DNS Record Type, Domain Name)
func CacheKey(msg *dns.Msg) string {
	var output string
	if len(msg.Answer) != 0 {
		output = fmt.Sprintf("%d:%s", msg.Answer[0].Header().Rrtype, msg.Answer[0].Header().Name)
	} else {
		output = fmt.Sprintf("%d:%s", msg.Question[0].Qtype, msg.Question[0].Name)
	}
	return output
}

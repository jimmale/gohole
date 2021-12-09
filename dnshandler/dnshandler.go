package dnshandler

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/ReneKroon/ttlcache/v2"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var hostRegex = regexp.MustCompile(`^0\.0\.0\.0 .*`) // TODO make this regex better
var infiniteDuration = time.Hour * 24 * 365 * 150

// GoholeHandler is kind of a proxy to the GoholeResolver
type GoholeHandler struct {
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

	// These are the domains that are blocked
	blockedDomains map[string]interface{}
	redactDomains  bool

	DNSCache *ttlcache.Cache
}

// NewGoholeResolver sets up a new Resolver
func NewGoholeResolver(c *cli.Context) *GoholeResolver {
	output := GoholeResolver{}

	log.Debugf("Creating GoholeResolver at %p", &output)

	output.blockedDomains = make(map[string]interface{})
	output.DNSCache = ttlcache.NewCache()
	output.DNSCache.SetLoaderFunction(output.getLoaderFunction())
	output.DNSCache.SetExpirationReasonCallback(output.getExpireCallbackFunction())

	// Make the cache return the _remaining_ ttl instead of the original ttl
	output.DNSCache.SkipTTLExtensionOnHit(true)

	if c != nil {
		// Set the upstream DNS
		output.UpstreamDNS = c.StringSlice("upstreamDNS")

		// Block all domains on the blocklists
		blocklists := c.StringSlice("blocklists")
		for index, blocklistURL := range blocklists {
			log.Debugf("Downloading blocklist %d of %d: %s", index+1, len(blocklists), blocklistURL)
			blockListContent, err := output.FetchBlocklist(blocklistURL)
			if err != nil {
				log.Errorf("Error when downloading %s: %s", blocklistURL, err.Error())
			} else {
				log.Tracef("Downloaded blocklist %d of %d: %s", index+1, len(blocklists), blocklistURL)
			}
			output.ApplyBlocklist(blockListContent)
		}

		log.Debugf("%d domains blocked from blocklists", len(output.blockedDomains))

		// Block all individual domains
		for _, v := range c.StringSlice("block") {
			output.BlockDomain(v)
		}
		log.Debugf("%d domains blocked individually", len(c.StringSlice("block")))

		// Unblock all individual domains
		// TODO

		output.redactDomains = !c.Bool("noredact")
	}
	return &output
}

// Resolve resolves a DNS query and returns a result.
func (ghr *GoholeResolver) Resolve(r *dns.Msg) *dns.Msg {
	log.Tracef("Resolving %s", ghr.redactDomain(r.Question[0].Name))
	cacheKey := CacheKey(r)                                 // calculate the cache key (eg "1:example.com." is the key for the A record for example.com)
	cacheEntry, ttl, _ := ghr.DNSCache.GetWithTTL(cacheKey) // fetch from cache; or load from recursive resolver
	msg := cacheEntry.(*dns.Msg)                            // cast to the appropriate struct

	savedRcode := msg.Rcode // save the Rcode because the dns.Msg.SetReply method sets it to 0 for some reason

	msg.SetReply(r)
	msg.Rcode = savedRcode // restore the Rcode
	msg.MsgHdr.RecursionAvailable = true
	msg.MsgHdr.Authoritative = true

	// the TTL of the message needs to be set to the actual expiration time
	if len(msg.Answer) > 0 {
		msg.Answer[0].Header().Ttl = uint32(ttl.Truncate(time.Second).Seconds())
	}

	return msg
}

// BlockDomain blocks a domain in the Resolver. All queries for that domain will return NXDomain.
func (ghr *GoholeResolver) BlockDomain(domain string) {
	ghr.blockedDomains[domain+"."] = true
}

// ApplyBlocklist applies the block list content to the resolver. the content must be in hosts file format
func (ghr *GoholeResolver) ApplyBlocklist(blocklistContent []byte) {
	// Example file format:
	/*
		#
		# This is a comment line
		#This is also a comment line
		0.0.0.0 blockedDomain.com
		12.34.56.78 examplewithip.com
		12.34.56.78 examplewithip2.com #comment
		12.34.56.78 examplewithip3.net#comment
	*/

	// TODO this probably doesn't parse all of the examples above
	sc := bufio.NewScanner(bytes.NewReader(blocklistContent))
	for sc.Scan() {
		line := strings.Trim(sc.Text(), "\t ")
		if hostRegex.Match([]byte(line)) {
			parts := strings.Split(line, " ")
			ghr.BlockDomain(parts[1])
		}
	}
}

// getExpireCallbackFunction gets the function that will be called when a cache item expires
func (ghr *GoholeResolver) getExpireCallbackFunction() func(key string, reason ttlcache.EvictionReason, value interface{}) {
	return func(key string, reason ttlcache.EvictionReason, cacheEntry interface{}) {
		msg := cacheEntry.(*dns.Msg)
		domain := msg.Answer[0].Header().Name
		log.Tracef("Entry for %s has expired", ghr.redactDomain(domain))
	}
}

// getLoaderFunction gets the function that will be called in the event of a cache miss
func (ghr *GoholeResolver) getLoaderFunction() func(string) (data interface{}, ttl time.Duration, err error) {

	// This loader function is a function that is called on a cache miss in order to check to see if the domain is
	//blocked, and if not, do a recursive resolution
	return func(key string) (data interface{}, ttl time.Duration, err error) {
		parts := strings.Split(key, ":")
		Qtypeint, _ := strconv.Atoi(parts[0])
		Qtypeuint16 := uint16(Qtypeint)
		domain := parts[1]

		log.Tracef("Could not find %s in cache.", ghr.redactDomain(domain))

		if _, domainIsBlocked := ghr.blockedDomains[domain]; domainIsBlocked {
			log.Tracef("Domain %s is blocked", ghr.redactDomain(domain))
			NXDomainMessage := new(dns.Msg)
			NXDomainMessage.Rcode = dns.RcodeNameError // NXDomain error
			NXDomainMessage.RecursionAvailable = true
			NXDomainMessage.Authoritative = true

			return NXDomainMessage, infiniteDuration, nil
		}

		log.Tracef("Domain %s is not blocked", ghr.redactDomain(domain))

		upstreamResponse := recursivelyResolve(Qtypeuint16, domain)
		newTTL := time.Duration(upstreamResponse.Answer[0].Header().Ttl)*time.Second - 1*time.Second

		log.Tracef("Entry for %s expires at %s", ghr.redactDomain(domain), time.Now().Add(newTTL).Format("2006.01.02 15:04:05.000 Z0700"))
		return upstreamResponse, newTTL, nil
	}
}

// recursivelyResolve makes an upstream DNS query.
func recursivelyResolve(recordType uint16, domain string) *dns.Msg {
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

// FetchBlocklist downloads the given block list and returns it as a string. Returns an error if downloading fails
func (ghr *GoholeResolver) FetchBlocklist(url string) ([]byte, error) {

	customDialer := &net.Dialer{
		Resolver: &net.Resolver{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Minute,
				}
				return d.DialContext(ctx, "udp", ghr.UpstreamDNS[0]+":53")
			},
			PreferGo: true,
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return customDialer.DialContext(ctx, network, addr)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: dialContext,
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gohole")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s returned HTTP status code %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// redactDomain will return either the provided string or [REDACTED] depending on the GoholeResolver setting
func (ghr *GoholeResolver) redactDomain(domain string) string {
	if ghr.redactDomains {
		return "[REDACTED]"
	} else {
		return domain
	}
}

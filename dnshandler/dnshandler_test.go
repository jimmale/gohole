package dnshandler

import (
	"github.com/ReneKroon/ttlcache/v2"
	"github.com/miekg/dns"
	"testing"
)

//func TestGoholeHandler_ServeDNS(t *testing.T) {
//	type fields struct {
//		blocklist map[string]bool
//		Resolver  *GoholeResolver
//	}
//	type args struct {
//		w dns.ResponseWriter
//		r *dns.Msg
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ghh := GoholeHandler{
//				blocklist: tt.fields.blocklist,
//				Resolver:  tt.fields.Resolver,
//			}
//		})
//	}
//}

// TestGoholeResolver_BlockDomainToCache tests to see if blocking a domain results in an entry being added to the cache
func TestGoholeResolver_BlockDomainToCache(t *testing.T) {
	// Setup
	myCache := ttlcache.NewCache()

	myResolver := GoholeResolver{
		UpstreamDNS:              nil,
		IndividualBlockedDomains: nil,
		IndividualAllowedDomains: nil,
		Blocklists:               nil,
		DNSCache:                 myCache,
	}

	// Precondition
	if _, err := myCache.Get("google.com."); err == nil {
		t.Error("fresh GoholeResolver contained an entry for google.com")
	}

	// Action
	myResolver.BlockDomain("google.com")

	// Postcondition

	// Fetching from the cache returns the correct response
	val, err := myCache.Get("google.com.")
	if err != nil {
		t.Error("a GoholeResolver did not hold on to a record")
	}

	result, ok := val.(*dns.Msg)
	if !ok {
		t.Error("casting interface{} to dns.Msg didn't work")
	}

	if result.Rcode != dns.RcodeNameError {
		t.Errorf("Cached DNS entry did not have NXDomain bit set")
	}
}

// TestGoholeResolver_BlockDomainToResolver checks to make sure that a blocked domain results in the resolver returning
// an NXDomain record with the correct DNS Packet ID.
func TestGoholeResolver_BlockDomainToResolver(t *testing.T) {
	// Setup
	myQuery := new(dns.Msg)
	myQuery.Id = dns.Id()
	myQuery.Question = make([]dns.Question, 1)
	myQuery.Question[0] = dns.Question{
		Name:   "google.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}

	myCache := ttlcache.NewCache()

	myResolver := GoholeResolver{
		UpstreamDNS:              nil,
		IndividualBlockedDomains: nil,
		IndividualAllowedDomains: nil,
		Blocklists:               nil,
		DNSCache:                 myCache,
	}

	myResolver.BlockDomain("google.com")

	// Action
	result := myResolver.Resolve(myQuery)

	// Postcondition
	if result.Id != myQuery.Id {
		t.Errorf("Returned query ID (%d) didn't match the original query (%d)", result.Id, myQuery.Id)
	}
	if result.Rcode != dns.RcodeNameError {
		t.Errorf("The blocked and cached DNS response didn't include the NX bit")
	}
}

// TestGoHoleResolver_DontConfuseRecordTypes makes sure that a cached DNS record of one type will not be returned for a
// query for a record of a different type. ie, Returning a cached AAAA record when the client requested an MX record
// would result in problems.
func TestGoHoleResolver_DontConfuseRecordTypes(t *testing.T) {
	// Setup
	QueryForMXRecord := new(dns.Msg)
	QueryForMXRecord.Id = dns.Id()
	QueryForMXRecord.Question = make([]dns.Question, 1)
	QueryForMXRecord.Question[0] = dns.Question{
		Name:   "google.com.",
		Qtype:  dns.TypeMX,
		Qclass: dns.ClassINET,
	}

	myCache := ttlcache.NewCache()

	myResolver := GoholeResolver{
		UpstreamDNS:              nil,
		IndividualBlockedDomains: nil,
		IndividualAllowedDomains: nil,
		Blocklists:               nil,
		DNSCache:                 myCache,
	}

	myRR, err := dns.NewRR("google.com. 3600 IN AAAA 2607:f8b0:4009:81c::200e")
	if err != nil {
		t.Error(err.Error())
	}

	cachedAAAARecord := new(dns.Msg)
	cachedAAAARecord.Rcode = dns.RcodeSuccess
	cachedAAAARecord.Answer = make([]dns.RR, 1)
	cachedAAAARecord.Answer[0] = myRR

	cacheKey := CacheKey(cachedAAAARecord)

	myCache.Set(cacheKey, cachedAAAARecord)

	// Action
	MXRecord := myResolver.Resolve(QueryForMXRecord)

	// Postcondition
	if MXRecord == nil {
		t.Error("MXRecord returned nil instead of NXDomain")
	}

	if len(MXRecord.Answer) > 0 {
		if MXRecord.Answer[0].Header().Rrtype != dns.TypeMX {
			t.Errorf("Requested an MX Record from the cache, but got an AAAAA record instead")
		}
	}
}

func TestRecursivelyResolve(t *testing.T) {
	QueryForMXRecord := new(dns.Msg)
	QueryForMXRecord.Id = dns.Id()
	QueryForMXRecord.Question = make([]dns.Question, 1)
	QueryForMXRecord.Question[0] = dns.Question{
		Name:   "google.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}

	myResolver := GoholeResolver{
		UpstreamDNS:              nil,
		IndividualBlockedDomains: nil,
		IndividualAllowedDomains: nil,
		Blocklists:               nil,
		DNSCache:                 nil,
	}

	response := myResolver.recursivelyResolve(QueryForMXRecord)
	if response == nil {
		t.Errorf("didn't get a response ")
	}

}

//
//func TestGoholeResolver_Resolve(t *testing.T) {
//	type fields struct {
//		UpstreamDNS              []string
//		IndividualBlockedDomains map[string]interface{}
//		IndividualAllowedDomains map[string]interface{}
//		Blocklists               []string
//		DNSCache                 *ttlcache.Cache
//	}
//	type args struct {
//		r *dns.Msg
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   dns.Msg
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ghr := &GoholeResolver{
//				UpstreamDNS:              tt.fields.UpstreamDNS,
//				IndividualBlockedDomains: tt.fields.IndividualBlockedDomains,
//				IndividualAllowedDomains: tt.fields.IndividualAllowedDomains,
//				Blocklists:               tt.fields.Blocklists,
//				DNSCache:                 tt.fields.DNSCache,
//			}
//			if got := ghr.Resolve(tt.args.r); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("Resolve() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestGoholeResolver_recursivelyResolve(t *testing.T) {
//	type fields struct {
//		UpstreamDNS              []string
//		IndividualBlockedDomains map[string]interface{}
//		IndividualAllowedDomains map[string]interface{}
//		Blocklists               []string
//		DNSCache                 *ttlcache.Cache
//	}
//	type args struct {
//		originalMessage *dns.Msg
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   *dns.Msg
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ghr := &GoholeResolver{
//				UpstreamDNS:              tt.fields.UpstreamDNS,
//				IndividualBlockedDomains: tt.fields.IndividualBlockedDomains,
//				IndividualAllowedDomains: tt.fields.IndividualAllowedDomains,
//				Blocklists:               tt.fields.Blocklists,
//				DNSCache:                 tt.fields.DNSCache,
//			}
//			if got := ghr.recursivelyResolve(tt.args.originalMessage); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("recursivelyResolve() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestNewGoholeResolver(t *testing.T) {
//	type args struct {
//		c *cli.Context
//	}
//	tests := []struct {
//		name string
//		args args
//		want *GoholeResolver
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := NewGoholeResolver(tt.args.c); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("NewGoholeResolver() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

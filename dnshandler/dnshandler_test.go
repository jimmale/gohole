package dnshandler

import (
	"github.com/miekg/dns"
	"testing"
)

// TestGoholeResolver_BlockDomainToResolver checks to make sure that a blocked domain results in the resolver returning
// an NXDomain record with the correct DNS Packet ID.
func TestGoholeResolver_BlockDomainToResolver(t *testing.T) {
	// Setup
	myQuery := new(dns.Msg)
	myQuery.Id = dns.Id()
	myQuery.Question = make([]dns.Question, 1)
	myQuery.Question[0] = dns.Question{
		Name:   "example.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}

	myResolver := NewGoholeResolver(nil)

	myResolver.BlockDomain("example.com")

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
// query for a record of a different type. ie, Returning a cached AAAA record when the client requested an A record
// would result in problems.
func TestGoHoleResolver_DontConfuseRecordTypes(t *testing.T) {
	// Setup
	QueryForARecord := new(dns.Msg)
	QueryForARecord.Id = dns.Id()
	QueryForARecord.Question = make([]dns.Question, 1)
	QueryForARecord.Question[0] = dns.Question{
		Name:   "example.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}
	myResolver := NewGoholeResolver(nil)

	// Cache an AAAA record, which should *not* be returned
	myRR, _ := dns.NewRR("example.com. 3600 IN AAAA 2001:db8::1")
	cachedAAAARecord := new(dns.Msg)
	cachedAAAARecord.Rcode = dns.RcodeSuccess
	cachedAAAARecord.Answer = make([]dns.RR, 1)
	cachedAAAARecord.Answer[0] = myRR
	cacheKey := CacheKey(cachedAAAARecord)
	_ = myResolver.DNSCache.Set(cacheKey, cachedAAAARecord)

	// Action
	ARecord := myResolver.Resolve(QueryForARecord)

	// Postcondition
	if ARecord == nil {
		t.Error("ARecord returned nil instead of NXDomain")
	}

	if ARecord.Answer[0].Header().Rrtype != dns.TypeA {
		t.Errorf("Requested an A Record from the cache, but got an AAAAA record instead")
	}
}

// TestRecursivelyResolve tests to make sure that a cache miss results in a recursive resolution
func TestRecursivelyResolve(t *testing.T) {
	QueryForARecord := new(dns.Msg)
	QueryForARecord.Id = dns.Id()
	QueryForARecord.Question = make([]dns.Question, 1)
	QueryForARecord.Question[0] = dns.Question{
		Name:   "example.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}

	myResolver := NewGoholeResolver(nil)

	response := myResolver.Resolve(QueryForARecord)
	if response == nil {
		t.Errorf("didn't get a response")
	}

	if response.Id != QueryForARecord.Id {
		t.Errorf("Returned query ID (%d) didn't match the original query (%d)", response.Id, QueryForARecord.Id)
	}
}

// TestResolveFromCache tests to make sure that a cache resolution results in the correct DNS exchange id being set on
// the second DNS response
func TestResolveFromCache(t *testing.T) {
	FirstQueryForARecord := new(dns.Msg)
	FirstQueryForARecord.Id = 42 // DNS Exchange ID
	FirstQueryForARecord.Question = make([]dns.Question, 1)
	FirstQueryForARecord.Question[0] = dns.Question{
		Name:   "example.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}

	SecondQueryForARecord := new(dns.Msg)
	SecondQueryForARecord.Id = 1000 // DNS Exchange ID
	SecondQueryForARecord.Question = make([]dns.Question, 1)
	SecondQueryForARecord.Question[0] = dns.Question{
		Name:   "example.com.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}

	myResolver := NewGoholeResolver(nil)

	firstResponse := myResolver.Resolve(FirstQueryForARecord)

	if firstResponse.Id != FirstQueryForARecord.Id {
		t.Errorf("First Query: Returned query ID (%d) didn't match the original query (%d)", firstResponse.Id, FirstQueryForARecord.Id)
	}

	secondResponse := myResolver.Resolve(SecondQueryForARecord)

	if secondResponse.Id != SecondQueryForARecord.Id {
		t.Errorf("Second Query: Returned query ID (%d) didn't match the original query (%d)", secondResponse.Id, SecondQueryForARecord.Id)
	}
}

// TestNewGoholeResolver tests to make sure that a new GoHole resolver is configured correctly
func TestNewGoholeResolver(t *testing.T) {
	//TODO
}

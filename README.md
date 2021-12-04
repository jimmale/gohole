# gohole
A GoLang DNS-based ad blocker

### Minimum Testable Product
- [x] Block a hard coded domain with NXDomain. For all others, return a record with 127.0.0.1

### Minimum Viable Product
- [x] Make DNS Requests to upstream DNS server
- [x] Reply to DNS Requests
- [x] Selectively return NXDomain based on blocklists

### TODO
- [ ] Configuration File
  - [ ] Debug logging on or off
  - [ ] List blocklists
  - [ ] List individually blocked domains
  - [ ] List 
- [ ] Download blocklists
- [ ] Configurable allowLists
- [ ] Cache DNS Requests
- [ ] Cache eviction
- [ ] Refresh blocklists
- [ ] systemd unit
- [ ] deb/rpm files
- [ ] Github Builds
- [ ] Metrics
  - Number of Blocked requests vs Number of Allowed Requests
  - Number of Requests served from cache vs Number of fresh requests
  - Duration of DNS responses based off of blocked/cached/resolved

Here's a good list of BlockLists:

https://github.com/StevenBlack/hosts
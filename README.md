# gohole
A GoLang DNS-based ad blocker

### Minimum Testable Product
- [ ] Block a hard coded domain with NXDomain. For all others, return a record with 127.0.0.1

### TODO
- [ ] Configuration File
- [ ] Download blocklists
- [ ] Make DNS Requests to upstream DNS server
- [ ] Reply to DNS Requests
- [ ] Selectively return NXDomain based on blocklists
- [ ] Configurable AllowLists
- [ ] Cache DNS Requests
- [ ] Cache eviction
- [ ] Identify records to keep hot
- [ ] Refresh blocklists
- [ ] deb/rpm files
- [ ] systemd unit
- [ ] Metrics
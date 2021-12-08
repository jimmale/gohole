# gohole
A GoLang DNS-based ad blocker


## Supported Platforms

| OS      | 386 | amd64 | arm6 | arm64 |
| ---     | --- | ----  | ---  | ----  |
| Linux   |     | ✅     | ✅    | ✅     |
| Windows |     |       |      |       |
| MacOS   |     |       |      |       |

## Building
### Requirements:
- GoLang 1.16+
- [Goreleaser](https://goreleaser.com/) (optional, to build linux packages)

### Build instructions
```
goreleaser release --rm-dist --snapshot
```

### Installing
TODO



### Building from source
```shell
go build .
```

### Configuration
Here's a good list of BlockLists:

https://github.com/StevenBlack/hosts
```
TODO
```

### Running in Docker
```shell
TODO
```

## Dependencies
### Buildtime
| Library                                                         | License | Purpose                                 |
| -------                                                         | ------- | -------                                 |
| [Sirupsen/Logrus](https://github.com/Sirupsen/logrus)           | MIT     | Pretty Logging                          |
| [BurntSushi/toml](https://github.com/BurntSushi/toml)           | MIT     | Config File Parsing                     |
| [urfave/cli](https://github.com/urfave/cli)                     | MIT     | Command line parameter management       |
| [miekg/dns](https://github.com/miekg/dns)                       | BSD-3   | DNS stuff                               |
| [ReneKroon/ttlcache](https://github.com/ReneKroon/ttlcache)     | MIT     | Caching DNS responses for faster lookup | 

### Runtime
- Working CA Certificate store (see [here](https://stackoverflow.com/a/40051432)) to download blocklists
- Systemd is recommended, but you can absolutely run it without.

## TODO
- [ ] Configuration File
  - [ ] Debug logging on or off
  - [ ] List blocklists
  - [ ] List individually blocked domains
  - [ ] Allowlist domains
  - [ ] Allowlist regex?
  - [ ] Upstream DNS servers
  - [ ] Easy blocklist set up (eg use Steve Black's list by default)
  - [ ] Easy configuration with env vars & [Steven Black's lists](https://github.com/StevenBlack/hosts)
- [ ] DNS over TLS for Android Private DNS
  - [ ] watch (inotify?) for updated TLS certs, automatically reload
- [ ] Download blocklists
  - [ ] Use configured upstream DNS to make these requests 
- [ ] Configurable allowLists
- [ ] Cache DNS Requests
  - [ ] Cache eviction
- [ ] Refresh blocklists
- [ ] systemd unit
  - [ ] PID File
  - [ ] Reload / SIGHUP handler
  - [ ] dbus liveness check
  - [ ] journald integration?
- [ ] deb package
  - [ ] cacert dependency
  - [ ] systemd recommendation
- [ ] rpm package
  - [ ] cacert dependency
  - [ ] systemd recommendation
- [ ] Github Builds
- [ ] Metrics
  - Number of Blocked requests vs Number of Allowed Requests
  - Number of Requests served from cache vs Number of fresh requests
  - Duration of DNS responses based off of blocked/cached/resolved

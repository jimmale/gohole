# gohole
A GoLang DNS-based ad blocker

## Supported Platforms

| OS      | 386 | amd64 | arm6 | arm64 |
| ---     | --- | ----  | ---  | ----  |
| Linux   |     | ✅     | ✅    | ✅     |
| Windows |     |       |      |       |
| MacOS   |     |       |      |       |

## Usage
### Installing
TODO

### Configuration
```toml
## /etc/gohole/gohole.toml

debug = true      # give some debug logging
trace = false     # give verbose debug logging
noredact = false  # set to "true" to show domain names in the logs

# Set your upstream DNS servers here
upstreamDNS = [
    # Cloudflare DNS
    "1.1.1.1",
    "1.0.0.1",

    #Google DNS
    # "8.8.8.8",
    # "8.8.4.4",
]

# List blocklists here
# a good source of lists is here: https://github.com/StevenBlack/hosts
blocklists = [
    "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-gambling/hosts",  # adware, malware, fakenews, gambling
]

# Block individual domains here
block = [
    "evilcompany.com",      # Evil Company gives us ads and viruses :(
    "www.evilcompany.com",
]

# Allow individual domains here
allow = [
    "en.wikipedia.org", # Wikipedia is a great source of information!
]
```

### Running as a service
```shell
TODO
```

### Running in Docker
```shell
TODO
```

### Invocation
```text
$ ./gohole --help
NAME:
   gohole - A GoLang DNS-based ad blocker

USAGE:
   gohole [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                 debug logging (default: false) [$GOHOLE_DEBUG]
   --trace, -v, --verbose  trace logging (default: false) [$GOHOLE_TRACE]
   --noredact              do not redact domain names in logs (default: false) [$GOHOLE_NOREDACT]
   --blocklists value      use blocklists (host file format) [$GOHOLE_BLOCKLIST]
   --block value           block individual domains [$GOHOLE_BLOCK]
   --allow value           allow individual domains [$GOHOLE_ALLOW]
   --upstreamDNS value     list upstream DNS servers to use (default: "1.1.1.1", "1.0.0.1") [$GOHOLE_UPSTREAMDNS]
   --config value          use a configuration file (default: "/etc/gohole/gohole.toml") [$GOHOLE_CONFIG_FILE]
   --help, -h              show help (default: false)
```

## Building
### Requirements:
- GoLang 1.16+
- [Goreleaser](https://goreleaser.com/) (optional, to build linux packages)

### Building from source
```shell
go test ./...
go build .
```

### Build packages
```shell
goreleaser release --rm-dist --snapshot
```


## Dependencies
### Buildtime
| Library                                                     | License | Purpose                                      |
| -------                                                     | ------- | -------                                      |
| [Sirupsen/Logrus](https://github.com/Sirupsen/logrus)       | MIT     | Pretty Logging                               |
| [urfave/cli/v2](https://github.com/urfave/cli/v2)            | MIT     | CLI args parsing, configuration file reading |
| [miekg/dns](https://github.com/miekg/dns)                   | BSD-3   | DNS stuff                                    |
| [ReneKroon/ttlcache](https://github.com/ReneKroon/ttlcache) | MIT     | Caching DNS responses for faster lookup      |


### Runtime
- Working CA Certificate store (see [here](https://stackoverflow.com/a/40051432)) to download blocklists
- Systemd is recommended, but you can absolutely run it without.

## TODO
- [ ] Embed build information
- [ ] Embed dependency licenses
- [ ] Use configured upstream DNS servers
  - [x] Config file entry
  - [ ] Actually use them
- [ ] Bind to interfaces better
  - [ ] Optionally List in config file
  - [ ] Bind to all by default
  - [ ] Don't bind to 127.0.0.53 to prevent clashing with systemd
- [ ] Parse hosts file format better
- [ ] move DNS client to a place where it can be reused
- [ ] DNS over TLS for Android Private DNS
  - [ ] watch (inotify?) for updated TLS certs, automatically reload
- [x] Download blocklists
  - [x] Use configured upstream DNS to make these requests 
  - [x] Set a User Agent
- [x] Cache DNS Requests
  - [x] Cache eviction
- [ ] Refresh blocklists periodically
- [ ] systemd unit
  - [ ] PID File
  - [ ] dbus liveness check
  - [ ] journald integration?
- [ ] Reload / SIGHUP handler
  - [ ] Reload configuration
  - [ ] Reload TLS certs
  - [ ] Purge DNS cache
  - [ ] Reload Blocklists
- [ ] deb package
  - [ ] cacert dependency
  - [ ] systemd recommendation
- [ ] rpm package
  - [ ] cacert dependency
  - [ ] systemd recommendation
- [ ] Github Builds
- [ ] Metrics
  - [ ] Number of Blocked requests vs Number of Allowed Requests
  - [ ] Cache hits vs cache misses (ttlcache provides this)
  - [ ] Duration of DNS responses based off of blocked/cached/resolved
  - [ ] Cache size (ignoring blocked hosts)
- [ ] Include license info for dependencies
- [x] Handle A, AAAA, and other records gracefully
  - ~~Right now a cached A record could prevent the successful resolution of AAAA, TXT, MX, etc.~~ 
- [ ] DNSSec
  - [ ] Validation
  - [ ] Forwarding
- [ ] Easy blocklist set up (eg use Steve Black's list by default)
- [ ] Easy configuration with env vars & [Steven Black's lists](https://github.com/StevenBlack/hosts)
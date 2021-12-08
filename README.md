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
Here's a good list of BlockLists:

https://github.com/StevenBlack/hosts
```
TODO
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
   --debug         debug logging (default: false) [$GOHOLE_DEBUG]
   --trace         trace logging (default: false) [$GOHOLE_TRACE]
   --noredact      do not redact domain names in logs (default: false) [$GOHOLE_NOREDACT]
   --config value  use a configuration file (default: "/etc/gohole/gohole.toml") [$GOHOLE_CONFIG_FILE]
   --help, -h      show help (default: false)
```

## Building
### Requirements:
- GoLang 1.16+
- [Goreleaser](https://goreleaser.com/) (optional, to build linux packages)

### Building from source
```shell
go build .
```

### Build packages
```
goreleaser release --rm-dist --snapshot
```


## Dependencies
### Buildtime
| Library                                                     | License | Purpose                                      |
| -------                                                     | ------- | -------                                      |
| [Sirupsen/Logrus](https://github.com/Sirupsen/logrus)       | MIT     | Pretty Logging                               |
| [urfave/cli/v2]("https://github.com/urfave/cli/v2")         | MIT     | CLI args parsing, configuration file reading |
| [miekg/dns](https://github.com/miekg/dns)                   | BSD-3   | DNS stuff                                    |
| [ReneKroon/ttlcache](https://github.com/ReneKroon/ttlcache) | MIT     | Caching DNS responses for faster lookup      |


### Runtime
- Working CA Certificate store (see [here](https://stackoverflow.com/a/40051432)) to download blocklists
- Systemd is recommended, but you can absolutely run it without.

## TODO
- [ ] Configuration File
  - [x] Debug logging on or off
  - [x] List blocklists
  - [x] Block individual domains
  - [x] Allow individual domains
  - [ ] Allowlist regex?
  - [x] Upstream DNS servers
  - [ ] Easy blocklist set up (eg use Steve Black's list by default)
  - [ ] Easy configuration with env vars & [Steven Black's lists](https://github.com/StevenBlack/hosts)
  - [ ] List of interfaces to bind to
- [ ] DNS over TLS for Android Private DNS
  - [ ] watch (inotify?) for updated TLS certs, automatically reload
- [ ] Download blocklists
  - [ ] Use configured upstream DNS to make these requests 
  - [ ] Set a User Agent
- [ ] Cache DNS Requests
  - [ ] Cache eviction
- [ ] Refresh blocklists periodically
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
- [ ] Include license info for dependencies
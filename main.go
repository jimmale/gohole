package main

import (
	"bufio"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)


func main() {

	var defaultDns = cli.NewStringSlice("1.1.1.1", "1.0.0.1")

	flags := []cli.Flag{
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "debug",
			Aliases:     nil,
			Usage:       "debug logging",
			EnvVars:     []string{"GOHOLE_DEBUG"},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			Value:       false,
			DefaultText: "",
			Destination: nil,
			HasBeenSet:  false,
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "trace",
			Aliases:     []string{"v", "verbose"},
			Usage:       "trace logging",
			EnvVars:     []string{"GOHOLE_TRACE"},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			Value:       false,
			DefaultText: "",
			Destination: nil,
			HasBeenSet:  false,
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "noredact",
			Aliases:     nil,
			Usage:       "do not redact domain names in logs",
			EnvVars:     []string{"GOHOLE_NOREDACT"},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			Value:       false,
			DefaultText: "",
			Destination: nil,
			HasBeenSet:  false,
		}),

		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name:        "blocklists",
			Aliases:     nil,
			Usage:       "use blocklists (host file format)",
			EnvVars:     []string{"GOHOLE_BLOCKLIST"},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       nil,
			DefaultText: "",
			HasBeenSet:  false,
			Destination: nil,
		}),

		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name:        "block",
			Aliases:     nil,
			Usage:       "block individual domains",
			EnvVars:     []string{"GOHOLE_BLOCK"},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       nil,
			DefaultText: "",
			HasBeenSet:  false,
			Destination: nil,
		}),

		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name:        "upstreamDNS",
			Aliases:     nil,
			Usage:       "list upstream DNS servers to use",
			EnvVars:     []string{"GOHOLE_UPSTREAMDNS"},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
		    Value:       defaultDns,
			DefaultText: "",
			HasBeenSet:  false,
			Destination: nil,
		}),

		&cli.StringFlag{
			Name:        "config",
			Aliases:     nil,
			Usage:       "use a configuration file",
			EnvVars:     []string{"GOHOLE_CONFIG_FILE"},
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       "/etc/gohole/gohole.toml",
			DefaultText: "",
			Destination: nil,
			HasBeenSet:  false,
		},
	}

	app := &cli.App{
		Name:   "gohole",
		Usage:  "A GoLang DNS-based ad blocker",
		Action: mainAction,
		Flags:  flags,
	}
	app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewTomlSourceFromFlagFunc("config"))
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func mainAction(c *cli.Context) error {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:               false,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             true,
		TimestampFormat:           "2006.01.02 15:04:05.000 Z0700",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    true,
		PadLevelText:              false,
		QuoteEmptyFields:          true,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	})

	// Set the log level
	switch {
	case c.Bool("trace"):
		{
			log.SetLevel(log.TraceLevel)
			log.Trace("Trace mode enabled")
		}
	case c.Bool("debug"):
		{
			log.SetLevel(log.DebugLevel)
			log.Debug("Debug mode enabled")
		}
	default:
		{
			log.SetLevel(log.InfoLevel)
		}
	}

	for k, v := range c.StringSlice("block"){
		log.Infof("%d %s", k, v)
	}

	//mh := MyHandler{}
	//mh.UpdateBlockList()
	//
	//log.Println("Ready.")
	//
	//bindAddr := getLocalIP() + ":53"
	//mh.UpdateBlockList()
	//log.Fatalf(dns.ListenAndServe(bindAddr, "udp4", mh).Error())
	return nil
}

type MyHandler struct {
	blocklist map[string]bool
}

func (m MyHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.MsgHdr.RecursionAvailable = true

	if m.DomainIsBlocked(r.Question[0].Name) {
		log.Debugf("%s is blocked", r.Question[0].Name)
		msg.Rcode = dns.RcodeNameError
	} else {
		log.Tracef("%s is not blocked", r.Question[0].Name)
		msg = Resolve(r)
	}
	w.WriteMsg(msg)
	w.Close()
}

func Resolve(originalMessage *dns.Msg) *dns.Msg {
	var output *dns.Msg

	// Make a recursive DNS call
	c := new(dns.Client)
	newMessage := new(dns.Msg)
	newMessage.SetQuestion(originalMessage.Question[0].Name, originalMessage.Question[0].Qtype)
	newMessage.RecursionDesired = true
	r, _, err := c.Exchange(newMessage, "1.1.1.1:53")

	if err != nil {
		log.Errorf("Error making recursive DNS Call: %s", err.Error())
	}

	output = r
	r.SetReply(originalMessage)

	return output
}

func (m *MyHandler) DomainIsBlocked(domain string) bool {
	_, blocked := m.blocklist[domain]
	return blocked
}

func (m *MyHandler) UpdateBlockList() {
	log.Println("Updating Blocklist")
	if m.blocklist == nil {
		log.Trace("blocklist is nil. Instantiating")
		m.blocklist = map[string]bool{}
	}

	reg := regexp.MustCompile(`^0\.0\.0\.0 .*`)

	resp, err := http.Get("BlockListURL")
	if err != nil {
		log.Errorf("Error fetching blocklist: %s", err.Error())
		return
	}
	defer resp.Body.Close()
	log.Trace("Downloaded Blocklist. Parsing...")

	sc := bufio.NewScanner(resp.Body)

	for sc.Scan() {
		line := sc.Text()
		if reg.Match([]byte(line)) {
			parts := strings.Split(line, " ")

			// domain names actually have a dot at the end.
			actualFQDN := parts[1] + "."
			m.blocklist[actualFQDN] = true
		}
	}
}

func getLocalIP() string {
	// TODO 127.0.0.53 is probably the only address we shouldn't try binding to.

	localIPRegex := regexp.MustCompile(`192\.168\..*`)
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, a := range addrs {
			ip, _, _ := net.ParseCIDR(a.String())
			if localIPRegex.Match([]byte(ip.To4().String())) {
				log.Printf("Local IP: %s", ip.To4().String())
				return ip.To4().String()
			}
		}
	}
	return ""
}

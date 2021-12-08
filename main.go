package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"os"
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

	for k, v := range c.StringSlice("block") {
		log.Infof("%d %s", k, v)
	}

	//mh := dnshandler.GoholeHandler{}
	//mh.UpdateBlockList()
	//
	//log.Println("Ready.")
	//
	//bindAddr := utils.GetLocalIP() + ":53"
	//mh.UpdateBlockList()
	//log.Fatalf(dns.ListenAndServe(bindAddr, "udp4", mh).Error())
	//return nil

	return nil
}

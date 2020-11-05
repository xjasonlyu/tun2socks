package cmd

import "github.com/urfave/cli/v2"

var (
	API = cli.StringFlag{
		Name:  "api",
		Usage: "URL of external API to listen",
	}

	Device = cli.StringFlag{
		Name:    "device",
		Aliases: []string{"d"},
		Usage:   "URL of device to open",
	}

	DNS = cli.StringFlag{
		Name:  "dns",
		Usage: "URL of fake DNS to listen",
	}

	Hosts = cli.StringSliceFlag{
		Name:  "hosts",
		Usage: "Extra hosts mapping",
	}

	Interface = cli.StringFlag{
		Name:    "interface",
		Aliases: []string{"i"},
		Usage:   "Bind interface to dial",
	}

	LogLevel = cli.StringFlag{
		Name:    "loglevel",
		Aliases: []string{"l"},
		Usage:   "Set logging level",
		Value:   "INFO",
	}

	Proxy = cli.StringFlag{
		Name:    "proxy",
		Aliases: []string{"p"},
		Usage:   "URL of proxy to dial",
	}

	Version = cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Print current version",
	}
)

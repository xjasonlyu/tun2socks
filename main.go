package main

import (
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/xjasonlyu/tun2socks/internal/cmd"
	"github.com/xjasonlyu/tun2socks/pkg/log"
)

const Usage = "A tun2socks implementation written in Go."

var (
	Version   string
	BuildTime string
)

func main() {
	app := &cli.App{
		Usage:   Usage,
		Version: Version,
		Action:  cmd.Main,
		Flags: []cli.Flag{
			&cmd.API,
			&cmd.Device,
			&cmd.DNS,
			&cmd.Hosts,
			&cmd.Interface,
			&cmd.LogLevel,
			&cmd.Proxy,
			&cmd.Version,
		},
		HideVersion:     true,
		HideHelpCommand: true,
	}
	app.Compiled, _ = time.Parse(time.RFC3339, BuildTime)

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}

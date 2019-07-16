// +build v2ray

package main

import (
	"context"
	"flag"
	"io/ioutil"
	"strings"

	vcore "v2ray.com/core"
	vproxyman "v2ray.com/core/app/proxyman"
	vbytespool "v2ray.com/core/common/bytespool"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/proxy/v2ray"
)

func init() {
	args.addFlag(fUdpTimeout)

	args.VConfig = flag.String("vconfig", "config.json", "Config file for v2ray, in JSON format, and note that routing in v2ray could not violate routes in the routing table")
	args.SniffingType = flag.String("sniffingType", "http,tls", "Enable domain sniffing for specific kind of traffic in v2ray")

	registerHandlerCreator("v2ray", func() {
		core.SetBufferPool(vbytespool.GetPool(core.BufSize))

		configBytes, err := ioutil.ReadFile(*args.VConfig)
		if err != nil {
			log.Fatalf("invalid vconfig file")
		}
		var validSniffings []string
		sniffings := strings.Split(*args.SniffingType, ",")
		for _, s := range sniffings {
			if s == "http" || s == "tls" {
				validSniffings = append(validSniffings, s)
			}
		}

		v, err := vcore.StartInstance("json", configBytes)
		if err != nil {
			log.Fatalf("start V instance failed: %v", err)
		}

		sniffingConfig := &vproxyman.SniffingConfig{
			Enabled:             true,
			DestinationOverride: validSniffings,
		}
		if len(validSniffings) == 0 {
			sniffingConfig.Enabled = false
		}

		ctx := vproxyman.ContextWithSniffingConfig(context.Background(), sniffingConfig)

		core.RegisterTCPConnHandler(v2ray.NewTCPHandler(ctx, v))
		core.RegisterUDPConnHandler(v2ray.NewUDPHandler(ctx, v, *args.UdpTimeout))
	})
}

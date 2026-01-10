package engine

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/v2/proxy"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy/direct"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy/http"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy/reject"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy/relay"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy/shadowsocks"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy/socks4"
	_ "github.com/xjasonlyu/tun2socks/v2/proxy/socks5"
)

func parseProxy(s string) (proxy.Proxy, error) {
	if !strings.Contains(s, "://") {
		s = fmt.Sprintf("%s://%s", "socks5" /* default protocol */, s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	return proxy.Parse(u)
}

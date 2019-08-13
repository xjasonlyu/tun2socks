package utils

import (
	"net"
	"strconv"
	"time"

	"github.com/xjasonlyu/tun2socks/common/cache"
	"github.com/xjasonlyu/tun2socks/common/lsof"
)

var (
	c *cache.Cache

	t = 120 * time.Second
)

func init() {
	c = cache.New(t)
}

func lookup(key interface{}) (string, bool) {
	item := c.Get(key)
	if item != nil {
		return item.(string), true
	}
	return "", false
}

func GetProcessName(addr net.Addr) string {
	if process, ok := lookup(addr); ok {
		return process
	}

	localHost, localPortStr, _ := net.SplitHostPort(addr.String())
	localPortInt, _ := strconv.Atoi(localPortStr)
	process, _ := lsof.GetCommandNameBySocket(addr.Network(), localHost, uint16(localPortInt))

	// put to cache
	c.Put(addr, process, t)
	return process
}

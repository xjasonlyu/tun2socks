package tunnel

import (
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

var (
	_globalMu sync.RWMutex
	_globalT  *Tunnel
)

func init() {
	ReplaceGlobal(New(&proxy.Base{}, statistic.DefaultManager))
	T().ProcessAsync()
}

// T returns the global Tunnel, which can be reconfigured with
// ReplaceGlobal. It's safe for concurrent use.
func T() *Tunnel {
	_globalMu.RLock()
	t := _globalT
	_globalMu.RUnlock()
	return t
}

// ReplaceGlobal replaces the global Tunnel, and returns a function
// to restore the original values. It's safe for concurrent use.
func ReplaceGlobal(t *Tunnel) func() {
	_globalMu.Lock()
	prev := _globalT
	_globalT = t
	_globalMu.Unlock()
	return func() { ReplaceGlobal(prev) }
}

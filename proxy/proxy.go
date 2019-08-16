package proxy

import (
	"errors"
	"net"

	D "github.com/xjasonlyu/tun2socks/component/fakedns"
	S "github.com/xjasonlyu/tun2socks/component/session"
)

var (
	fakeDNS D.FakeDNS
	monitor S.Monitor
)

func RegisterFakeDNS(d D.FakeDNS) {
	fakeDNS = d
}

func RegisterMonitor(m S.Monitor) {
	monitor = m
}

// DNS lookup
func lookupHost(target net.Addr) (targetHost string, err error) {
	var targetIP net.IP
	switch addr := target.(type) {
	case *net.TCPAddr:
		targetIP = addr.IP
	case *net.UDPAddr:
		targetIP = addr.IP
	default:
		err = errors.New("invalid target type")
		return
	}

	targetHost = targetIP.String()
	// Replace with a domain name if target address IP is a fake IP.
	if fakeDNS != nil {
		if host, exist := fakeDNS.IPToHost(targetIP); exist {
			targetHost = host
		}
	}
	return
}

// Session Operation
func addSession(key interface{}, session *S.Session) {
	if monitor != nil {
		monitor.AddSession(key, session)
	}
}

func removeSession(key interface{}) {
	if monitor != nil {
		monitor.RemoveSession(key)
	}
}

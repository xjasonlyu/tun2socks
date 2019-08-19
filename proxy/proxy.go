package proxy

import (
	"errors"
	"net"
	"strconv"
	"strings"

	D "github.com/xjasonlyu/tun2socks/component/fakedns"
	S "github.com/xjasonlyu/tun2socks/component/session"
)

var (
	monitor S.Monitor

	fakeDNS D.FakeDNS
	// default DNS address
	hijackDNS []string
)

// Register Monitor
func RegisterMonitor(m S.Monitor) {
	monitor = m
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

// Register FakeDNS
func RegisterFakeDNS(d D.FakeDNS, h string) {
	fakeDNS = d
	hijackDNS = append(hijackDNS, strings.Split(h, ",")...)
}

// Check target if is hijacked address.
func isHijacked(target *net.UDPAddr) bool {
	for _, addr := range hijackDNS {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			continue
		}
		portInt, _ := strconv.Atoi(port)
		if (host == "*" && portInt == target.Port) || addr == target.String() {
			return true
		}
	}
	return false
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

package dns

import (
	"net"
)

type FakeDNS interface {
	Start() error
	Stop() error
	// IPToHost returns the corresponding domain for the given IP.
	IPToHost(ip net.IP) (string, bool)
}

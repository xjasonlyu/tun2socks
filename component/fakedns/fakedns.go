package fakedns

import (
	"net"
)

type FakeDNS interface {
	// Resolve a fake dns response for the specify request.
	Resolve([]byte) ([]byte, error)

	// IPToHost returns the corresponding domain for the given IP.
	IPToHost(ip net.IP) (string, bool)
}

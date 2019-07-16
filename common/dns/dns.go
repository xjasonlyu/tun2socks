package dns

import (
	"net"
)

const CommonDnsPort = 53

type FakeDns interface {
	// GenerateFakeResponse generates a fake dns response for the specify request.
	// GenerateFakeResponse(request []byte) ([]byte, error)

	// IPToHost returns the corresponding domain for the given IP.
	IPToHost(ip net.IP) string

	// IsFakeIP checks if the given ip is a fake IP.
	// IsFakeIP(ip net.IP) bool
}

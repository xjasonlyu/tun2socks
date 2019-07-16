package dns

import (
	"net"
)

const CommonDnsPort = 53

type DnsCache interface {
	// Query queries the response for the DNS request with payload `p`,
	// the response data should be a valid DNS response payload.
	Query(p []byte) []byte

	// Store stores the DNS response with payload `p` to the cache.
	Store(p []byte)
}

type FakeDns interface {
	// GenerateFakeResponse generates a fake dns response for the specify request.
	// GenerateFakeResponse(request []byte) ([]byte, error)

	// IPToHost returns the corresponding domain for the given IP.
	IPToHost(ip net.IP) string

	// IsFakeIP checks if the given ip is a fake IP.
	IsFakeIP(ip net.IP) bool
}

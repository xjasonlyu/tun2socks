package socks4

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsReservedIP(t *testing.T) {
	reservedIPs := []string{
		"0.0.0.1",
		"0.0.0.2",
		"0.0.0.50",
		"0.0.0.100",
		"0.0.0.255",
	}
	for _, ip := range reservedIPs {
		assert.True(t, isReservedIP(netip.MustParseAddr(ip)))
	}

	unReservedIPs := []string{
		"0.0.0.0",
		"0.0.1.0",
		"1.1.1.1",
		"10.0.0.0",
		"255.255.255.255",
	}
	for _, ip := range unReservedIPs {
		assert.False(t, isReservedIP(netip.MustParseAddr(ip)))
	}
}

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		addr string
		host string
		port uint16
	}{
		{
			"1.1.1.1:80",
			"1.1.1.1",
			80,
		},
		{
			"1.1.1.1:0",
			"1.1.1.1",
			0,
		},
		{
			"0.0.0.0:0",
			"0.0.0.0",
			0,
		},
		{
			"[::1]:443",
			"::1",
			443,
		},
		{
			"example.com:80",
			"example.com",
			80,
		},
	}
	for _, tt := range tests {
		host, port, err := splitHostPort(tt.addr)
		assert.NoError(t, err)
		assert.Equal(t, tt.host, host)
		assert.Equal(t, tt.port, port)
	}

	addrs := []string{
		"1.1.1.1:-80",
		"1.1.1.1:abcd",
		"::1:80",
		"[::1]",
		"example.com",
	}
	for _, addr := range addrs {
		_, _, err := splitHostPort(addr)
		assert.Error(t, err)
	}
}

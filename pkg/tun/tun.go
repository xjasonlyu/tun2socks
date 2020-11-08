package tun

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const defaultScheme = "tun"

type Device interface {
	stack.LinkEndpoint

	Name() string // returns the current name
	Close() error // stops and closes the tun
}

// Open opens TUN Device with given URL.
func Open(rawURL string) (Device, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = defaultScheme + "://" + rawURL
	}

	var u *url.URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(u.Scheme) != defaultScheme {
		return nil, errors.New("unsupported TUN scheme")
	}

	var n uint64
	if mtu := u.Query().Get("mtu"); mtu != "" {
		n, _ = strconv.ParseUint(mtu, 10, 32)
	}

	name := u.Host
	return CreateTUN(name, uint32(n))
}

package dev

import (
	"errors"
	"io"
	"net/url"
	"strings"

	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/internal/dev/tun"
)

const defaultScheme = "tun"

type Device struct {
	url *url.URL
	io.Closer
	stack.LinkEndpoint
}

func Open(deviceURL string) (device *Device, err error) {
	if !strings.Contains(deviceURL, "://") {
		deviceURL = defaultScheme + "://" + deviceURL
	}

	var u *url.URL
	if u, err = url.Parse(deviceURL); err != nil {
		return
	}

	var (
		ep stack.LinkEndpoint
		c  io.Closer
	)
	switch strings.ToLower(u.Scheme) {
	case "tun":
		name := u.Host
		ep, c, err = tun.Open(name)
	default:
		err = errors.New("unsupported device type")
	}

	if err != nil {
		return
	}

	device = &Device{
		url:          u,
		Closer:       c,
		LinkEndpoint: ep,
	}
	return
}

// Close closes device.
func (d *Device) Close() error {
	return d.Closer.Close()
}

// Name returns name of device.
func (d *Device) Name() string {
	return d.url.Host
}

// Type returns type of device.
func (d *Device) Type() string {
	return strings.ToLower(d.url.Scheme)
}

// String returns full URL string.
func (d *Device) String() string {
	return d.url.String()
}

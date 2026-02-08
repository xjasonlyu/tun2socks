//go:build !unix && !windows

package dialer

import (
	"net"
)

func WithBindToInterface(_ *net.Interface) SocketOption { return UnsupportedSocketOption }

func WithRoutingMark(_ int) SocketOption { return UnsupportedSocketOption }

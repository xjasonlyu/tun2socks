package core

import (
	"errors"
	"net"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"

	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/core/option"
)

const (
	// defaultWndSize if set to zero, the default
	// receive window buffer size is used instead.
	defaultWndSize = 0

	// maxConnAttempts specifies the maximum number
	// of in-flight tcp connection attempts.
	maxConnAttempts = 2 << 10

	// tcpKeepaliveCount is the maximum number of
	// TCP keep-alive probes to send before giving up
	// and killing the connection if no response is
	// obtained from the other end.
	tcpKeepaliveCount = 9

	// tcpKeepaliveIdle specifies the time a connection
	// must remain idle before the first TCP keepalive
	// packet is sent. Once this time is reached,
	// tcpKeepaliveInterval option is used instead.
	tcpKeepaliveIdle = 60 * time.Second

	// tcpKeepaliveInterval specifies the interval
	// time between sending TCP keepalive packets.
	tcpKeepaliveInterval = 30 * time.Second
)

func withTCPHandler(handle func(adapter.TCPConnSYN), printf func(string, ...any)) option.Option {
	return func(s *stack.Stack) error {
		tcpForwarder := tcp.NewForwarder(s, defaultWndSize, maxConnAttempts, func(r *tcp.ForwarderRequest) {
			var (
				err tcpip.Error
				id  = r.ID()
			)

			defer func() {
				if err != nil {
					printf("forward tcp request %s:%d->%s:%d: %s",
						id.RemoteAddress, id.RemotePort, id.LocalAddress, id.LocalPort, err)
				}
			}()

			conn := &tcpConnSYN{
				id: id,
				r:  r,
				s:  s,
			}
			handle(conn)
		})
		s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)
		return nil
	}
}

func setSocketOptions(s *stack.Stack, ep tcpip.Endpoint) tcpip.Error {
	{ /* TCP keepalive options */
		ep.SocketOptions().SetKeepAlive(true)

		idle := tcpip.KeepaliveIdleOption(tcpKeepaliveIdle)
		if err := ep.SetSockOpt(&idle); err != nil {
			return err
		}

		interval := tcpip.KeepaliveIntervalOption(tcpKeepaliveInterval)
		if err := ep.SetSockOpt(&interval); err != nil {
			return err
		}

		if err := ep.SetSockOptInt(tcpip.KeepaliveCountOption, tcpKeepaliveCount); err != nil {
			return err
		}
	}
	{ /* TCP recv/send buffer size */
		var ss tcpip.TCPSendBufferSizeRangeOption
		if err := s.TransportProtocolOption(header.TCPProtocolNumber, &ss); err == nil {
			ep.SocketOptions().SetReceiveBufferSize(int64(ss.Default), false)
		}

		var rs tcpip.TCPReceiveBufferSizeRangeOption
		if err := s.TransportProtocolOption(header.TCPProtocolNumber, &rs); err == nil {
			ep.SocketOptions().SetReceiveBufferSize(int64(rs.Default), false)
		}
	}
	return nil
}

type tcpConnSYN struct {
	id stack.TransportEndpointID
	r  *tcp.ForwarderRequest
	s  *stack.Stack
}

func (c *tcpConnSYN) ID() *stack.TransportEndpointID {
	return &c.id
}

func (c *tcpConnSYN) CompleteHandshake() (net.Conn, error) {
	// Perform a TCP three-way handshake.
	r := c.r
	s := c.s
	var wq waiter.Queue
	var ep tcpip.Endpoint
	ep, err := r.CreateEndpoint(&wq)
	if err != nil {
		// RST: prevent potential half-open TCP connection leak.
		r.Complete(true)
		return nil, errors.New(err.String())
	}
	defer r.Complete(false)

	err = setSocketOptions(s, ep)
	if err != nil {
		r.Complete(true)
		return nil, errors.New(err.String())
	}

	tcpConn := gonet.NewTCPConn(&wq, ep)
	return tcpConn, nil
}

func (c *tcpConnSYN) RST() {
	c.r.Complete(true)
}

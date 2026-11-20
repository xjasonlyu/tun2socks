package core

import (
	"context"
	"errors"
	"net"

	glog "gvisor.dev/gvisor/pkg/log"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"

	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/core/option"
)

func withDeferredTCPHandler(connector adapter.TCPConnector, cfg DeferredTCPConfig, rejectIPv6 bool) option.Option {
	return func(s *stack.Stack) error {
		ctx := cfg.Context
		failures := newTCPFailureCache(cfg.FailureCacheTTL, cfg.FailureCacheCapacity)
		context.AfterFunc(ctx, failures.stop)
		f := newTrackedTCPForwarder(s, defaultWndSize, maxConnAttempts, func(r *tcp.ForwarderRequest, key tcpFailureKey, complete func(bool)) {
			// Exactly one completion per SYN, including every early error path.
			completed := false
			defer func() {
				if !completed {
					complete(true)
				}
			}()
			var (
				wq  waiter.Queue
				ep  tcpip.Endpoint
				err tcpip.Error
				id  = r.ID()
			)

			defer func() {
				if err != nil {
					glog.Debugf("forward tcp request: %s:%d->%s:%d: %s",
						id.RemoteAddress, id.RemotePort, id.LocalAddress, id.LocalPort, err)
				}
			}()

			if ctx.Err() != nil {
				return
			}
			var remote net.Conn
			var stopRemoteWatch func() bool
			transferred := false
			defer func() {
				if !transferred && remote != nil {
					if stopRemoteWatch != nil {
						stopRemoteWatch()
					}
					remote.Close()
				}
			}()
			{
				if failures.contains(key) {
					return // Same failed SYN: reset again without another proxy dial.
				}
				var dialErr error
				remote, dialErr = connector.DialTCP(ctx, id)
				if dialErr != nil || remote == nil {
					if ctx.Err() == nil && !errors.Is(dialErr, context.Canceled) {
						failures.add(key)
					}
					glog.Debugf("forward tcp dial: %s:%d: %v", id.LocalAddress, id.LocalPort, dialErr)
					return // RST in response to SYN, not a close after successful connect.
				}
				stopRemoteWatch = context.AfterFunc(ctx, func() { remote.Close() })
				if ctx.Err() != nil {
					return
				}
			}

			// Only now advertise success to the application (send SYN-ACK).
			ep, err = r.CreateEndpoint(&wq)
			if err != nil {
				return
			}
			complete(false)
			completed = true

			err = setSocketOptions(s, ep)
			if err != nil {
				ep.Close()
				return
			}

			conn := &deferredTCPConn{
				TCPConn: gonet.NewTCPConn(&wq, ep),
				id:      id,
			}
			conn.stopWatch = context.AfterFunc(ctx, func() { conn.TCPConn.Close() })
			if ctx.Err() != nil {
				conn.Close()
				return
			}
			{
				// Close cancels both lifetime watches after the handler finishes.
				conn.stopRemoteWatch = stopRemoteWatch
				transferred = true
				connector.HandleTCPConnected(conn, remote)
			}
		})
		return setTransportHandler(s, tcp.ProtocolNumber, f.HandlePacket, rejectIPv6)
	}
}

type deferredTCPConn struct {
	*gonet.TCPConn
	id              stack.TransportEndpointID
	stopWatch       func() bool
	stopRemoteWatch func() bool
}

func (c *deferredTCPConn) Close() error {
	if c.stopWatch != nil {
		c.stopWatch()
	}
	if c.stopRemoteWatch != nil {
		c.stopRemoteWatch()
	}
	return c.TCPConn.Close()
}

func (c *deferredTCPConn) ID() stack.TransportEndpointID {
	return c.id
}

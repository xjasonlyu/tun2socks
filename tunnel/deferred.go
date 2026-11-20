package tunnel

import (
	"context"
	"errors"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/log"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

// DeferredTCPHandler is an explicit adapter for core.Config.DeferredTCP.
// It does not change Tunnel's public API, queues or default lifecycle. The caller
// owns its lifetime: cancel the constructor context or call Close before closing
// the associated stack. Closing it does not close the wrapped Tunnel.
type DeferredTCPHandler struct {
	tunnel   *Tunnel
	ctx      context.Context
	cancel   context.CancelFunc
	timeouts proxy.TCPTimeouts
}

var (
	_ adapter.TransportHandler = (*DeferredTCPHandler)(nil)
	_ adapter.TCPConnector     = (*DeferredTCPHandler)(nil)
)

func NewDeferredTCPHandler(ctx context.Context, t *Tunnel) *DeferredTCPHandler {
	h, _ := NewDeferredTCPHandlerWithConfig(ctx, t, proxy.TCPTimeouts{})
	return h // Zero settings always normalize successfully.
}

// NewDeferredTCPHandlerWithConfig selects dialing timeouts without changing the
// core TCPConnector interface or the Tunnel's legacy/default behavior.
func NewDeferredTCPHandlerWithConfig(ctx context.Context, t *Tunnel, timeouts proxy.TCPTimeouts) (*DeferredTCPHandler, error) {
	timeouts, err := timeouts.WithDefaults()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	return &DeferredTCPHandler{tunnel: t, ctx: ctx, cancel: cancel, timeouts: timeouts}, nil
}

func (h *DeferredTCPHandler) Close()                      { h.cancel() }
func (h *DeferredTCPHandler) HandleTCP(c adapter.TCPConn) { h.tunnel.HandleTCP(c) }
func (h *DeferredTCPHandler) HandleUDP(c adapter.UDPConn) { h.tunnel.HandleUDP(c) }

func tcpMetadata(id stack.TransportEndpointID) *M.Metadata {
	return &M.Metadata{
		Network: M.TCP, SrcIP: parseTCPIPAddress(id.RemoteAddress), SrcPort: id.RemotePort,
		DstIP: parseTCPIPAddress(id.LocalAddress), DstPort: id.LocalPort,
	}
}

// DialTCP includes the proxy CONNECT result, not merely a connection to its port.
func (h *DeferredTCPHandler) DialTCP(parent context.Context, id stack.TransportEndpointID) (net.Conn, error) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	stop := context.AfterFunc(h.ctx, cancel)
	defer stop()
	metadata := tcpMetadata(id)
	p := h.tunnel.Proxy()
	var remote net.Conn
	var err error
	if bounded, ok := p.(interface {
		DialContextWithTimeouts(context.Context, *M.Metadata, proxy.TCPTimeouts) (net.Conn, error)
	}); ok {
		remote, err = bounded.DialContextWithTimeouts(ctx, metadata, h.timeouts)
	} else {
		// Other proxy implementations must honor cancellation throughout
		// DialContext, not just when connecting to the proxy's port.
		dialCtx, cancelDial := context.WithTimeout(ctx, tcpConnectTimeout)
		defer cancelDial()
		remote, err = p.DialContext(dialCtx, metadata)
		if err == nil {
			err = dialCtx.Err()
		}
	}
	if err == nil {
		err = ctx.Err()
	}
	if err == nil && remote == nil {
		err = errors.New("proxy returned no TCP connection")
	}
	if err != nil {
		if remote != nil {
			remote.Close()
		}
		log.Warnf("[TCP] dial %s: %v", metadata.DestinationAddress(), err)
		return nil, err
	}
	return remote, nil
}

func (h *DeferredTCPHandler) HandleTCPConnected(origin adapter.TCPConn, remote net.Conn) {
	if h.ctx.Err() != nil {
		origin.Close()
		remote.Close()
		return
	}
	go h.relay(origin, remote)
}

func (h *DeferredTCPHandler) relay(origin adapter.TCPConn, remote net.Conn) {
	defer origin.Close()
	defer remote.Close()
	stop := context.AfterFunc(h.ctx, func() { origin.Close(); remote.Close() })
	defer stop()
	metadata := tcpMetadata(origin.ID())
	metadata.MidIP, metadata.MidPort = parseNetAddr(remote.LocalAddr())
	tracked := statistic.NewTCPTracker(remote, metadata, h.tunnel.manager)
	defer tracked.Close()
	log.Infof("[TCP] %s <-> %s", metadata.SourceAddress(), metadata.DestinationAddress())
	pipe(origin, tracked)
}

package dialer

import (
	"context"
	"net"
	"sync"
	"syscall"

	"go.uber.org/atomic"
)

// DefaultDialer is the package-level default Dialer.
// It is used by DialContext and ListenPacket.
var DefaultDialer = &Dialer{}

// RegisterSockOpt registers a socket option on the DefaultDialer.
func RegisterSockOpt(opt SocketOption) {
	DefaultDialer.RegisterSockOpt(opt)
}

// DialContext dials using the DefaultDialer.
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return DefaultDialer.DialContext(ctx, network, address)
}

// ListenPacket listens using the DefaultDialer.
func ListenPacket(network, address string) (net.PacketConn, error) {
	return DefaultDialer.ListenPacket(network, address)
}

// Dialer applies registered SocketOptions to all dials/listens.
type Dialer struct {
	optsMu     sync.Mutex
	atomicOpts atomic.Value
}

// New creates a new Dialer with the given initial socket options.
func New(opts ...SocketOption) *Dialer {
	d := &Dialer{}
	for _, opt := range opts {
		d.RegisterSockOpt(opt)
	}
	return d
}

// RegisterSockOpt registers a socket option on the Dialer.
func (d *Dialer) RegisterSockOpt(opt SocketOption) {
	d.optsMu.Lock()
	opts, _ := d.atomicOpts.Load().([]SocketOption)
	d.atomicOpts.Store(append(opts, opt))
	d.optsMu.Unlock()
}

func (d *Dialer) applySockOpts(network string, address string, c syscall.RawConn) error {
	opts, _ := d.atomicOpts.Load().([]SocketOption)
	if len(opts) == 0 {
		return nil
	}
	// Skip non-global-unicast IPs (e.g. loopback, link-local).
	if host, _, err := net.SplitHostPort(address); err == nil {
		if ip := net.ParseIP(host); ip != nil && !ip.IsGlobalUnicast() {
			return nil
		}
	}
	for _, opt := range opts {
		if err := opt.Apply(network, address, c); err != nil {
			return err
		}
	}
	return nil
}

// DialContext behaves like net.Dialer.DialContext, applying registered SocketOptions.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	nd := &net.Dialer{
		Control: d.applySockOpts,
	}
	return nd.DialContext(ctx, network, address)
}

// ListenPacket behaves like net.ListenConfig.ListenPacket, applying registered SocketOptions.
func (d *Dialer) ListenPacket(network, address string) (net.PacketConn, error) {
	lc := &net.ListenConfig{
		Control: d.applySockOpts,
	}
	return lc.ListenPacket(context.Background(), network, address)
}

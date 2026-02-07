package dialer

import (
	"context"
	"net"
	"sync"
	"syscall"

	"go.uber.org/atomic"
)

// DefaultDialer is the default Dialer and is used by DialContext and ListenPacket.
var DefaultDialer = &Dialer{}

// RegisterSockOpt registers the given socket option to DefaultDialer.
func RegisterSockOpt(opt SocketOption) {
	DefaultDialer.RegisterSockOpt(opt)
}

// DialContext is a wrapper around DefaultDialer.DialContext.
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return DefaultDialer.DialContext(ctx, network, address)
}

// ListenPacket is a wrapper around DefaultDialer.ListenPacket.
func ListenPacket(network, address string) (net.PacketConn, error) {
	return DefaultDialer.ListenPacket(network, address)
}

type Dialer struct {
	optsMu     sync.Mutex
	atomicOpts atomic.Value
}

func New(opts ...SocketOption) *Dialer {
	d := &Dialer{}
	for _, opt := range opts {
		d.RegisterSockOpt(opt)
	}
	return d
}

func (d *Dialer) RegisterSockOpt(opt SocketOption) {
	d.optsMu.Lock()
	opts, _ := d.atomicOpts.Load().([]SocketOption)
	d.atomicOpts.Store(append(opts, opt))
	d.optsMu.Unlock()
}

func (d *Dialer) applySockOpt(network string, address string, c syscall.RawConn) error {
	host, _, _ := net.SplitHostPort(address)
	if ip := net.ParseIP(host); ip != nil && !ip.IsGlobalUnicast() {
		return nil
	}
	opts, _ := d.atomicOpts.Load().([]SocketOption)
	for _, opt := range opts {
		if err := opt.Apply(network, address, c); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return (&net.Dialer{
		Control: d.applySockOpt,
	}).DialContext(ctx, network, address)
}

func (d *Dialer) ListenPacket(network, address string) (net.PacketConn, error) {
	return (&net.ListenConfig{
		Control: d.applySockOpt,
	}).ListenPacket(context.Background(), network, address)
}

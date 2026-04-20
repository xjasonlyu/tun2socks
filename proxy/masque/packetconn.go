package masque

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"

	"github.com/xjasonlyu/tun2socks/v2/log"
)

// h3DatagramConn adapts an *http3.RequestStream carrying RFC 9298
// context-ID-0 HTTP/3 datagrams to a net.PacketConn.
type h3DatagramConn struct {
	rs     *http3.RequestStream
	ctx    context.Context
	cancel context.CancelFunc
	target *net.UDPAddr

	mu         sync.Mutex // protects readDl, writeDl, readCancel
	readDl     time.Time
	writeDl    time.Time
	readCancel context.CancelFunc // cancels the in-flight ReceiveDatagram, if any

	closeOnce sync.Once
	closeErr  error
	done      chan struct{}
}

var _ net.PacketConn = (*h3DatagramConn)(nil)

// WriteTo prepends the RFC 9298 §5 context ID (0 = raw UDP payload) as a
// QUIC varint and sends the result as an HTTP/3 datagram. The destination
// address is ignored: the stream is already bound to a single target by
// the CONNECT-UDP request. SendDatagram queues internally and does not
// block, so checking the deadline before the call is sufficient.
func (c *h3DatagramConn) WriteTo(p []byte, _ net.Addr) (int, error) {
	c.mu.Lock()
	dl := c.writeDl
	c.mu.Unlock()
	if !dl.IsZero() && time.Now().After(dl) {
		return 0, os.ErrDeadlineExceeded
	}
	buf := quicvarint.Append(make([]byte, 0, len(p)+1), 0)
	buf = append(buf, p...)
	if err := c.rs.SendDatagram(buf); err != nil {
		var tooLarge *quic.DatagramTooLargeError
		if errors.As(err, &tooLarge) {
			return 0, fmt.Errorf("masque: datagram too large (max payload=%d, sent=%d): %w",
				tooLarge.MaxDatagramPayloadSize, len(buf), err)
		}
		return 0, err
	}
	return len(p), nil
}

// ReadFrom blocks for the next HTTP/3 datagram whose context ID is 0,
// strips the varint, and copies the payload into p. Datagrams with
// non-zero context IDs (reserved by RFC 9298) are silently dropped.
//
// SetReadDeadline can interrupt a blocked ReadFrom: it cancels the
// in-flight context, the loop observes context.Canceled, and re-derives
// a new context from the updated deadline.
func (c *h3DatagramConn) ReadFrom(p []byte) (int, net.Addr, error) {
	for {
		c.mu.Lock()
		dl := c.readDl
		var ctx context.Context
		var cancel context.CancelFunc
		if dl.IsZero() {
			ctx, cancel = context.WithCancel(c.ctx)
		} else {
			ctx, cancel = context.WithDeadline(c.ctx, dl)
		}
		c.readCancel = cancel
		c.mu.Unlock()

		data, err := c.rs.ReceiveDatagram(ctx)

		c.mu.Lock()
		c.readCancel = nil
		c.mu.Unlock()
		cancel()

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return 0, c.target, os.ErrDeadlineExceeded
			}
			if errors.Is(err, context.Canceled) {
				select {
				case <-c.done:
					return 0, c.target, err
				default:
					// Deadline was changed mid-call; loop and retry
					// with the new deadline.
					continue
				}
			}
			return 0, c.target, err
		}
		cid, n, perr := quicvarint.Parse(data)
		if perr != nil {
			continue // malformed datagram; drop per RFC 9298 §5
		}
		if cid != 0 {
			continue // reserved for future extensions
		}
		payload := data[n:]
		copied := copy(p, payload)
		if copied < len(payload) {
			// Match real UDP sockets: truncate silently rather than
			// returning a non-nil error, which tunnel/udp.go treats
			// as terminal and would tear the session down.
			log.Debugf("[MASQUE] datagram truncated (%d > %d)", len(payload), len(p))
		}
		return copied, c.target, nil
	}
}

func (c *h3DatagramConn) Close() error {
	c.closeOnce.Do(func() {
		c.cancel()
		close(c.done)
		c.mu.Lock()
		if c.readCancel != nil {
			c.readCancel()
		}
		c.mu.Unlock()
		// rs.Close() only sends FIN on the send side; the drainCapsules
		// goroutine reads from the receive side and would block until
		// the peer closes. CancelRead wakes it up.
		c.rs.CancelRead(quic.StreamErrorCode(h3RequestCancelled))
		c.closeErr = c.rs.Close()
	})
	return c.closeErr
}

func (c *h3DatagramConn) LocalAddr() net.Addr {
	return &net.UDPAddr{}
}

func (c *h3DatagramConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	c.readDl = t
	c.writeDl = t
	if c.readCancel != nil {
		c.readCancel()
	}
	c.mu.Unlock()
	return nil
}

func (c *h3DatagramConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	c.readDl = t
	if c.readCancel != nil {
		c.readCancel()
	}
	c.mu.Unlock()
	return nil
}

func (c *h3DatagramConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	c.writeDl = t
	c.mu.Unlock()
	return nil
}

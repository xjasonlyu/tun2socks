package masque

import (
	"errors"
	"io"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"

	"github.com/xjasonlyu/tun2socks/v2/log"
)

// drainCapsules consumes HTTP/3 capsules from the stream's byte channel.
// This is mandatory: the RequestStream multiplexes DATA-frame bytes on
// the underlying QUIC stream, and if no one reads them, QUIC stream-level
// flow control eventually stalls the whole connection. RFC 9298 permits
// unrecognised capsule types and requires skipping them.
func drainCapsules(pc *h3DatagramConn) {
	r := quicvarint.NewReader(pc.rs)
	for {
		ct, body, err := http3.ParseCapsule(r)
		if err != nil {
			if !errors.Is(err, io.EOF) && !isClosed(pc) {
				log.Debugf("[MASQUE] capsule drain ended: %v", err)
			}
			return
		}
		if _, err := io.Copy(io.Discard, body); err != nil {
			log.Debugf("[MASQUE] capsule 0x%x body discard: %v", uint64(ct), err)
			return
		}
		log.Debugf("[MASQUE] drained capsule type=0x%x", uint64(ct))
	}
}

func isClosed(pc *h3DatagramConn) bool {
	select {
	case <-pc.done:
		return true
	default:
		return false
	}
}

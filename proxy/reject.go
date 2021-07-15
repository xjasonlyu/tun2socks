package proxy

import (
	"context"
	"errors"
	"net"

	M "github.com/xjasonlyu/tun2socks/constant"
	"github.com/xjasonlyu/tun2socks/proxy/proto"
)

var _ Proxy = (*Reject)(nil)

type Reject struct {
	*Base
}

func NewReject() *Reject {
	return &Reject{
		Base: &Base{
			proto: proto.Reject,
		},
	}
}

func (r *Reject) DialContext(context.Context, *M.Metadata) (net.Conn, error) {
	return nil, errors.New("TCP rejected")
}

func (r *Reject) DialUDP(*M.Metadata) (net.PacketConn, error) {
	return nil, errors.New("UDP rejected")
}

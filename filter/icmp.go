package filter

import (
	"io"

	"github.com/xjasonlyu/tun2socks/common/packet"
	"github.com/xjasonlyu/tun2socks/log"
)

type icmpFilter struct {
	writer io.Writer
}

func NewICMPFilter(w io.Writer) Filter {
	return &icmpFilter{writer: w}
}

func (f *icmpFilter) Write(buf []byte) (int, error) {
	switch buf[9] {
	case packet.PROTOCOL_ICMP:
		payload := make([]byte, len(buf))
		copy(payload, buf)
		go func(data []byte) {
			if _, err := f.writer.Write(data); err != nil {
				log.Fatalf("failed to input data to the stack: %v", err)
			}
		}(payload)
		return len(buf), nil
	default:
		return f.writer.Write(buf)
	}
}

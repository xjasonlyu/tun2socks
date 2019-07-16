package filter

import (
	"io"
	"time"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/packet"
)

type icmpFilter struct {
	writer io.Writer
	delay  int
}

func NewICMPFilter(w io.Writer, delay int) Filter {
	return &icmpFilter{writer: w, delay: delay}
}

func (w *icmpFilter) Write(buf []byte) (int, error) {
	if uint8(buf[9]) == packet.PROTOCOL_ICMP {
		payload := make([]byte, len(buf))
		copy(payload, buf)
		go func(data []byte) {
			time.Sleep(time.Duration(w.delay) * time.Millisecond)
			_, err := w.writer.Write(data)
			if err != nil {
				log.Fatalf("failed to input data to the stack: %v", err)
			}
		}(payload)
		return len(buf), nil
	} else {
		return w.writer.Write(buf)
	}
}

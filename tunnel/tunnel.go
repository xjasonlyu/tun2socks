package tunnel

import (
	"github.com/xjasonlyu/tun2socks/v2/core"
)

// Unbuffered TCP/UDP queues.
var (
	_tcpQueue = make(chan core.TCPConn)
	_udpQueue = make(chan core.UDPConn)
)

func init() {
	go process()
}

// TCPIn return fan-in TCP queue.
func TCPIn() chan<- core.TCPConn {
	return _tcpQueue
}

// UDPIn return fan-in UDP queue.
func UDPIn() chan<- core.UDPConn {
	return _udpQueue
}

func process() {
	for {
		select {
		case conn := <-_tcpQueue:
			go handleTCPConn(conn)
		case conn := <-_udpQueue:
			go handleUDPConn(conn)
		}
	}
}

package tunnel

import (
	"runtime"

	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/log"
)

const (
	// maxUDPQueueSize is the max number of UDP packets
	// could be buffered. if queue is full, upcoming packets
	// would be dropped util queue is ready again.
	maxUDPQueueSize = 1 << 9
)

var (
	_tcpQueue      = make(chan core.TCPConn) /* unbuffered */
	_udpQueue      = make(chan core.UDPPacket, maxUDPQueueSize)
	_numUDPWorkers = max(runtime.NumCPU(), 4 /* at least 4 workers */)
)

func init() {
	go process()
}

// Add adds tcpConn to tcpQueue.
func Add(conn core.TCPConn) {
	_tcpQueue <- conn
}

// AddPacket adds udpPacket to udpQueue.
func AddPacket(packet core.UDPPacket) {
	select {
	case _udpQueue <- packet:
	default:
		log.Warnf("queue is currently full, packet will be dropped")
		packet.Drop()
	}
}

func process() {
	for i := 0; i < _numUDPWorkers; i++ {
		queue := _udpQueue
		go func() {
			for packet := range queue {
				handleUDP(packet)
			}
		}()
	}

	for conn := range _tcpQueue {
		go handleTCP(conn)
	}
}

package tunnel

import (
	"runtime"

	"github.com/xjasonlyu/tun2socks/common/adapter"
	"github.com/xjasonlyu/tun2socks/log"
)

const (
	// maxUDPQueueSize is the max number of UDP packets
	// could be buffered. if queue is full, upcoming packets
	// would be dropped util queue is ready again.
	maxUDPQueueSize = 2 << 10
)

var (
	tcpQueue      = make(chan adapter.TCPConn) /* unbuffered */
	udpQueue      = make(chan adapter.UDPPacket, maxUDPQueueSize)
	numUDPWorkers = max(runtime.NumCPU(), 4 /* at least 4 workers */)
)

func init() {
	go process()
}

// Add adds tcpConn to tcpQueue.
func Add(conn adapter.TCPConn) {
	tcpQueue <- conn
}

// AddPacket adds udpPacket to udpQueue.
func AddPacket(packet adapter.UDPPacket) {
	select {
	case udpQueue <- packet:
	default:
		log.Warnf("queue is currently full, packet will be dropped")
		packet.Drop()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func process() {
	for i := 0; i < numUDPWorkers; i++ {
		queue := udpQueue
		go func() {
			for packet := range queue {
				handleUDP(packet)
			}
		}()
	}

	for conn := range tcpQueue {
		go handleTCP(conn)
	}
}

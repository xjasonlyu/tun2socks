package tunnel

import (
	"runtime"

	"github.com/xjasonlyu/tun2socks/internal/adapter"
	"github.com/xjasonlyu/tun2socks/pkg/log"
)

const (
	// maxUDPQueueSize is the max number of UDP packets
	// could be buffered. if queue is full, upcoming packets
	// would be dropped util queue is ready again.
	maxUDPQueueSize = 2 << 10
)

var (
	numUDPWorkers = max(runtime.NumCPU(), 4 /* at least 4 workers */)

	tcpQueue      = make(chan adapter.TCPConn) /* unbuffered */
	udpMultiQueue = make([]chan adapter.UDPPacket, 0, numUDPWorkers)
)

func init() {
	for i := 0; i < numUDPWorkers; i++ {
		udpMultiQueue = append(udpMultiQueue, make(chan adapter.UDPPacket, maxUDPQueueSize))
	}

	go process()
}

// Add adds tcpConn to tcpQueue.
func Add(conn adapter.TCPConn) {
	tcpQueue <- conn
}

// AddPacket adds udpPacket to udpQueue.
func AddPacket(packet adapter.UDPPacket) {
	m := packet.Metadata()
	// In order to keep each packet sent in order, we
	// calculate which queue each packet should be sent
	// by src/dst info, and make sure the rest of them
	// would only be sent to the same queue.
	i := int(m.SrcPort+m.DstPort) % numUDPWorkers

	select {
	case udpMultiQueue[i] <- packet:
	default:
		log.Warnf("queue is currently full, packet will be dropped")
		packet.Drop()
	}
}

func process() {
	for _, udpQueue := range udpMultiQueue {
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

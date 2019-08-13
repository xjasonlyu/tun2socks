package redirect

import (
	"net"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/core"
	. "github.com/xjasonlyu/tun2socks/proxy/utils"
)

// To do a benchmark using iperf3 locally, you may follow these steps:
//
// 1. Setup and configure the TUN device and start tun2socks with the
//    redirect handler using the following command:
//      tun2socks -proxyType redirect -proxyServer 127.0.0.1:1234
//    Tun2socks will redirect all traffic to 127.0.0.1:1234.
//
// 2. Route traffic targeting 1.2.3.4 to the TUN interface (240.0.0.1):
//      route add 1.2.3.4/32 240.0.0.1
//
// 3. Run iperf3 server locally and listening on 1234 port:
//      iperf3 -s -p 1234
//
// 4. Run iperf3 client locally and connect to 1.2.3.4:1234:
//      iperf3 -c 1.2.3.4 -p 1234
//
// It works this way:
// iperf3 client -> 1.2.3.4:1234 -> routing table -> TUN (240.0.0.1) -> tun2socks -> tun2socks redirect anything to 127.0.0.1:1234 -> iperf3 server
//
type tcpHandler struct {
	target string
}

func NewTCPHandler(target string) core.TCPConnHandler {
	return &tcpHandler{target: target}
}

func (h *tcpHandler) Handle(localConn net.Conn, target *net.TCPAddr) error {
	remoteConn, err := net.Dial("tcp", h.target)
	if err != nil {
		return err
	}

	// set keepalive
	TCPKeepAlive(localConn)
	TCPKeepAlive(remoteConn)

	go TCPRelay(localConn, remoteConn)

	log.Infof("new proxy connection for target: %s:%s", target.Network(), target.String())
	return nil
}

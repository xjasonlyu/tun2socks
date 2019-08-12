package direct

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/common/pool"
	"github.com/xjasonlyu/tun2socks/core"
)

type udpHandler struct {
	timeout time.Duration

	remoteUDPConnMap sync.Map
}

func NewUDPHandler(timeout time.Duration) core.UDPConnHandler {
	return &udpHandler{timeout: timeout}
}

func (h *udpHandler) fetchUDPInput(conn core.UDPConn, pc *net.UDPConn) {
	buf := pool.BufPool.Get().([]byte)

	defer func() {
		h.Close(conn)
		pool.BufPool.Put(buf[:cap(buf)])
	}()

	for {
		pc.SetDeadline(time.Now().Add(h.timeout))
		n, addr, err := pc.ReadFromUDP(buf)
		if err != nil {
			if err, ok := err.(net.Error); !ok && !err.Timeout() {
				log.Warnf("failed to read UDP data from remote: %v", err)
			}
			return
		}

		if _, err := conn.WriteFrom(buf[:n], addr); err != nil {
			log.Warnf("failed to write UDP data to TUN: %v", err)
			return
		}
	}
}

func (h *udpHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	bindAddr := &net.UDPAddr{IP: nil, Port: 0}
	pc, err := net.ListenUDP("udp", bindAddr)
	if err != nil {
		log.Errorf("failed to bind udp address")
		return err
	}

	h.remoteUDPConnMap.Store(conn, pc)

	go h.fetchUDPInput(conn, pc)

	log.Infof("new proxy connection for target: %s:%s", target.Network(), target.String())
	return nil
}

func (h *udpHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	var pc *net.UDPConn

	if value, ok := h.remoteUDPConnMap.Load(conn); ok {
		pc = value.(*net.UDPConn)
	}

	if pc != nil {
		_, err := pc.WriteToUDP(data, addr)
		if err != nil {
			log.Warnf("failed to write UDP payload to SOCKS5 server: %v", err)
			return errors.New("failed to write UDP data")
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("proxy connection %v->%v does not exists", conn.LocalAddr(), addr))
	}
}

func (h *udpHandler) Close(conn core.UDPConn) {
	conn.Close()

	if pc, ok := h.remoteUDPConnMap.Load(conn); ok {
		pc.(*net.UDPConn).Close()
		h.remoteUDPConnMap.Delete(conn)
	}
}

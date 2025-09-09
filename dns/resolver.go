package dns

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	"github.com/xjasonlyu/tun2socks/v2/log"
)

var (
	dnsConfig     *Config
	dnsConfigOnce sync.Once
)

// Config holds DNS configuration
type Config struct {
	Hijack  bool
	Address string
}

// SetConfig sets the global DNS configuration
func SetConfig(cfg *Config) {
	dnsConfigOnce.Do(func() {
		dnsConfig = cfg
	})
}

// GetConfig returns the global DNS configuration
func GetConfig() *Config {
	return dnsConfig
}

// IsDNSRequest checks if the request is a DNS request (port 53)
func IsDNSRequest(port uint16) bool {
	return port == 53
}

// ForwardDNSOverTCP forwards DNS query over TCP to the configured DNS server
func ForwardDNSOverTCP(clientConn net.Conn, dstAddr string) error {
	if dnsConfig == nil || !dnsConfig.Hijack {
		return nil // DNS hijacking is disabled
	}

	// Connect to the DNS server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dnsConn, err := dialer.DefaultDialer.DialContext(ctx, "tcp", dnsConfig.Address)
	if err != nil {
		log.Warnf("[DNS-TCP] failed to connect to DNS server %s: %v", dnsConfig.Address, err)
		return err
	}
	defer dnsConn.Close()

	log.Infof("[DNS-TCP] %s <-> %s", clientConn.RemoteAddr(), dnsConfig.Address)

	// Copy data bidirectionally
	done := make(chan error, 2)

	go func() {
		_, err := io.Copy(dnsConn, clientConn)
		done <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, dnsConn)
		done <- err
	}()

	// Wait for either direction to complete
	<-done
	return nil
}

// ForwardDNSOverUDP forwards DNS query over UDP to the configured DNS server
func ForwardDNSOverUDP(clientConn net.PacketConn, clientAddr net.Addr, data []byte) error {
	if dnsConfig == nil || !dnsConfig.Hijack {
		return nil // DNS hijacking is disabled
	}

	// Connect to the DNS server
	dnsConn, err := net.Dial("udp", dnsConfig.Address)
	if err != nil {
		log.Warnf("[DNS-UDP] failed to connect to DNS server %s: %v", dnsConfig.Address, err)
		return err
	}
	defer dnsConn.Close()

	log.Debugf("[DNS-UDP] forwarding query from %s to %s", clientAddr, dnsConfig.Address)

	// Send query to DNS server
	if _, err := dnsConn.Write(data); err != nil {
		log.Warnf("[DNS-UDP] failed to send query to DNS server: %v", err)
		return err
	}

	// Set read timeout
	dnsConn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read response from DNS server
	response := make([]byte, 4096)
	n, err := dnsConn.Read(response)
	if err != nil {
		log.Warnf("[DNS-UDP] failed to read response from DNS server: %v", err)
		return err
	}

	// Send response back to client
	if _, err := clientConn.WriteTo(response[:n], clientAddr); err != nil {
		log.Warnf("[DNS-UDP] failed to send response to client: %v", err)
		return err
	}

	log.Debugf("[DNS-UDP] forwarded response from %s to %s", dnsConfig.Address, clientAddr)
	return nil
}

func init() {
	// We must use this DialContext to query DNS
	// when using net default resolver.
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = dialer.DefaultDialer.DialContext
}

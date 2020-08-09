package session

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Monitor interface {
	Start() error
	Stop() error

	// METHODS
	AddSession(key interface{}, session *Session)
	RemoveSession(key interface{})
}

type Status struct {
	Platform       string    `json:"platform"`
	Version        string    `json:"version"`
	CPU            string    `json:"cpu"`
	MEM            string    `json:"mem"`
	Uptime         string    `json:"uptime"`
	Total          int64     `json:"total"`
	Upload         int64     `json:"upload"`
	Download       int64     `json:"download"`
	Goroutines     int       `json:"goroutines"`
	ActiveSessions []Session `json:"activeSessions"`
	ClosedSessions []Session `json:"closedSessions"`
}

type Session struct {
	Process       string    `json:"process"`
	Network       string    `json:"network"`
	DialerAddr    string    `json:"dialerAddr"`
	ClientAddr    string    `json:"clientAddr"`
	TargetAddr    string    `json:"targetAddr"`
	UploadBytes   int64     `json:"upload"`
	DownloadBytes int64     `json:"download"`
	SessionStart  time.Time `json:"sessionStart"`
	SessionClose  time.Time `json:"sessionClose"`
}

// Track SessionConn
type Conn struct {
	*Session
	net.Conn
	once sync.Once
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 {
		atomic.AddInt64(&c.DownloadBytes, int64(n))
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 {
		atomic.AddInt64(&c.UploadBytes, int64(n))
	}
	return
}

func (c *Conn) Close() error {
	c.once.Do(func() {
		c.SessionClose = time.Now()
	})
	return c.Conn.Close()
}

// Track SessionPacketConn
type PacketConn struct {
	*Session
	net.PacketConn
	once sync.Once
}

func (c *PacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, addr, err = c.PacketConn.ReadFrom(b)
	if n > 0 {
		atomic.AddInt64(&c.DownloadBytes, int64(n))
	}
	return
}

func (c *PacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	n, err = c.PacketConn.WriteTo(b, addr)
	if n > 0 {
		atomic.AddInt64(&c.UploadBytes, int64(n))
	}
	return
}

func (c *PacketConn) Close() error {
	c.once.Do(func() {
		c.SessionClose = time.Now()
	})
	return c.PacketConn.Close()
}

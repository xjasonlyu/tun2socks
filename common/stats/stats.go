package stats

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type SessionStater interface {
	Start()
	Stop()
	AddSession(key interface{}, session *Session)
	GetSession(key interface{}) *Session
	RemoveSession(key interface{})
}

type Session struct {
	Process       string
	Network       string
	DialerAddr    string
	ClientAddr    string
	TargetAddr    string
	UploadBytes   int64
	DownloadBytes int64
	SessionStart  time.Time
	SessionClose  time.Time
}

// Track SessionConn
type SessionConn struct {
	net.Conn
	once    sync.Once
	session *Session
}

func NewSessionConn(conn net.Conn, session *Session) net.Conn {
	return &SessionConn{
		Conn:    conn,
		session: session,
	}
}

func (c *SessionConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 {
		atomic.AddInt64(&c.session.DownloadBytes, int64(n))
	}
	return
}

func (c *SessionConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 {
		atomic.AddInt64(&c.session.UploadBytes, int64(n))
	}
	return
}

func (c *SessionConn) Close() error {
	c.once.Do(func() {
		c.session.SessionClose = time.Now()
	})
	return c.Conn.Close()
}

// Track SessionPacketConn
type SessionPacketConn struct {
	net.PacketConn
	once    sync.Once
	session *Session
}

func NewSessionPacketConn(conn net.PacketConn, session *Session) net.PacketConn {
	return &SessionPacketConn{
		PacketConn: conn,
		session:    session,
	}
}

func (c *SessionPacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, addr, err = c.PacketConn.ReadFrom(b)
	if n > 0 {
		atomic.AddInt64(&c.session.DownloadBytes, int64(n))
	}
	return
}

func (c *SessionPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	n, err = c.PacketConn.WriteTo(b, addr)
	if n > 0 {
		atomic.AddInt64(&c.session.UploadBytes, int64(n))
	}
	return
}

func (c *SessionPacketConn) Close() error {
	c.once.Do(func() {
		c.session.SessionClose = time.Now()
	})
	return c.PacketConn.Close()
}

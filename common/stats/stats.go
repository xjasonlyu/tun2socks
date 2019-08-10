package stats

import (
	"net"
	"sync/atomic"
	"time"
)

type SessionStater interface {
	Start() error
	Stop() error
	AddSession(key interface{}, session *Session)
	GetSession(key interface{}) *Session
	RemoveSession(key interface{})
}

type Session struct {
	ProcessName   string
	Network       string
	DialerAddr    string
	ClientAddr    string
	TargetAddr    string
	UploadBytes   int64
	DownloadBytes int64
	SessionStart  time.Time
}

type SessionConn struct {
	net.Conn
	*Session
}

func NewSessionConn(conn net.Conn, session *Session) net.Conn {
	return &SessionConn{
		Conn:    conn,
		Session: session,
	}
}

func (c *SessionConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 {
		atomic.AddInt64(&c.DownloadBytes, int64(n))
	}
	return
}

func (c *SessionConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 {
		atomic.AddInt64(&c.UploadBytes, int64(n))
	}
	return
}

type SessionPacketConn struct {
	net.PacketConn
	*Session
}

func NewSessionPacketConn(conn net.PacketConn, session *Session) net.PacketConn {
	return &SessionPacketConn{
		PacketConn: conn,
		Session:    session,
	}
}

func (c *SessionPacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, addr, err = c.PacketConn.ReadFrom(b)
	if n > 0 {
		atomic.AddInt64(&c.DownloadBytes, int64(n))
	}
	return
}

func (c *SessionPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	n, err = c.PacketConn.WriteTo(b, addr)
	if n > 0 {
		atomic.AddInt64(&c.UploadBytes, int64(n))
	}
	return
}

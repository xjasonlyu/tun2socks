package statistic

import (
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"go.uber.org/atomic"

	M "github.com/TianHe-Labs/Zeus/metadata"
)

type tracker interface {
	ID() string
	Close() error
}

type trackerInfo struct {
	Start         time.Time     `json:"start"`
	UUID          uuid.UUID     `json:"id"`
	Metadata      *M.Metadata   `json:"metadata"`
	UploadTotal   *atomic.Int64 `json:"upload"`
	DownloadTotal *atomic.Int64 `json:"download"`
	IsHandled     bool          `json:"isHandled"`
}

type TcpTracker struct {
	net.Conn `json:"-"`

	*trackerInfo
	manager *Manager
}

func NewTcpTracker(conn net.Conn, metadata *M.Metadata, manager *Manager) net.Conn {
	id, _ := uuid.NewRandom()

	tt := &TcpTracker{
		Conn:    conn,
		manager: manager,
		trackerInfo: &trackerInfo{
			UUID:          id,
			Start:         time.Now(),
			Metadata:      metadata,
			UploadTotal:   atomic.NewInt64(0),
			DownloadTotal: atomic.NewInt64(0),
		},
	}

	manager.Join(tt)
	return tt
}

// DefaultTcpTracker returns a new net.Conn(*tcpTacker) with default manager.
func DefaultTcpTracker(conn net.Conn, metadata *M.Metadata) net.Conn {
	return NewTcpTracker(conn, metadata, DefaultManager)
}

func (tt *TcpTracker) ID() string {
	return tt.UUID.String()
}

func (tt *TcpTracker) Read(b []byte) (int, error) {
	n, err := tt.Conn.Read(b)
	download := int64(n)
	tt.manager.PushDownloaded(download)
	tt.DownloadTotal.Add(download)
	return n, err
}

func (tt *TcpTracker) Write(b []byte) (int, error) {
	n, err := tt.Conn.Write(b)
	upload := int64(n)
	tt.manager.PushUploaded(upload)
	tt.UploadTotal.Add(upload)
	return n, err
}

func (tt *TcpTracker) Close() error {
	tt.manager.Leave(tt)
	return tt.Conn.Close()
}

func (tt *TcpTracker) CloseRead() error {
	if cr, ok := tt.Conn.(interface{ CloseRead() error }); ok {
		return cr.CloseRead()
	}
	return errors.New("CloseRead is not implemented")
}

func (tt *TcpTracker) CloseWrite() error {
	if cw, ok := tt.Conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return errors.New("CloseWrite is not implemented")
}

type udpTracker struct {
	net.PacketConn `json:"-"`

	*trackerInfo
	manager *Manager
}

func NewUDPTracker(conn net.PacketConn, metadata *M.Metadata, manager *Manager) net.PacketConn {
	id, _ := uuid.NewRandom()

	ut := &udpTracker{
		PacketConn: conn,
		manager:    manager,
		trackerInfo: &trackerInfo{
			UUID:          id,
			Start:         time.Now(),
			Metadata:      metadata,
			UploadTotal:   atomic.NewInt64(0),
			DownloadTotal: atomic.NewInt64(0),
		},
	}

	manager.Join(ut)
	return ut
}

// DefaultUDPTracker returns a new net.PacketConn(*udpTacker) with default manager.
func DefaultUDPTracker(conn net.PacketConn, metadata *M.Metadata) net.PacketConn {
	return NewUDPTracker(conn, metadata, DefaultManager)
}

func (ut *udpTracker) ID() string {
	return ut.UUID.String()
}

func (ut *udpTracker) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := ut.PacketConn.ReadFrom(b)
	download := int64(n)
	ut.manager.PushDownloaded(download)
	ut.DownloadTotal.Add(download)
	return n, addr, err
}

func (ut *udpTracker) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := ut.PacketConn.WriteTo(b, addr)
	upload := int64(n)
	ut.manager.PushUploaded(upload)
	ut.UploadTotal.Add(upload)
	return n, err
}

func (ut *udpTracker) Close() error {
	ut.manager.Leave(ut)
	return ut.PacketConn.Close()
}

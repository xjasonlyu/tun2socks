package statistic

import (
	"net"
	"time"

	M "github.com/xjasonlyu/tun2socks/v2/constant"

	"github.com/gofrs/uuid"
	"go.uber.org/atomic"
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
}

type tcpTracker struct {
	net.Conn `json:"-"`

	*trackerInfo
	manager *Manager
}

func NewTCPTracker(conn net.Conn, metadata *M.Metadata, manager *Manager) *tcpTracker {
	id, _ := uuid.NewV4()

	tt := &tcpTracker{
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

func (tt *tcpTracker) ID() string {
	return tt.UUID.String()
}

func (tt *tcpTracker) Read(b []byte) (int, error) {
	n, err := tt.Conn.Read(b)
	download := int64(n)
	tt.manager.PushDownloaded(download)
	tt.DownloadTotal.Add(download)
	return n, err
}

func (tt *tcpTracker) Write(b []byte) (int, error) {
	n, err := tt.Conn.Write(b)
	upload := int64(n)
	tt.manager.PushUploaded(upload)
	tt.UploadTotal.Add(upload)
	return n, err
}

func (tt *tcpTracker) Close() error {
	tt.manager.Leave(tt)
	return tt.Conn.Close()
}

type udpTracker struct {
	net.PacketConn `json:"-"`

	*trackerInfo
	manager *Manager
}

func NewUDPTracker(conn net.PacketConn, metadata *M.Metadata, manager *Manager) *udpTracker {
	id, _ := uuid.NewV4()

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

//go:build unix

package sockbased

import (
	"errors"
	"net"
	"strconv"
	"syscall"

	"github.com/xjasonlyu/tun2socks/v2/core/device"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type FD struct {
	stack.LinkEndpoint

	fd  int
	mtu uint32
}

func Open(sockpath string, mtu uint32) (device.Device, error) {
	// unlink删除已存在的unixSock文件
	syscall.Unlink(sockpath)
	laddr, err := net.ResolveUnixAddr("unix", sockpath)
	if err != nil {
		return nil, err
	}
	l, err := net.ListenUnix("unix", laddr)
	if err != nil {
		return nil, err
	}
	conn, err := l.AcceptUnix()
	if err != nil {
		return nil, err
	}
	// msg分为两部分数据
	buf := make([]byte, 32)
	oob := make([]byte, 32)
	_, oobn, _, _, err := conn.ReadMsgUnix(buf, oob)
	if err != nil {
		return nil, err
	}
	conn.Close()
	// 解出SocketControlMessage数组
	scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, err
	}
	if len(scms) == 0 {
		return nil, errors.New("syscall.ParseSocketControlMessage() length == 0")
	}
	// 从SocketControlMessage中得到UnixRights
	fds, err := syscall.ParseUnixRights(&(scms[0]))
	if err != nil {
		return nil, err
	}
	fd := fds[0]
	ep, err := fdbased.New(&fdbased.Options{
		FDs: []int{fd},
		MTU: mtu,
		// TUN only, ignore ethernet header.
		EthernetHeader: false,
	})
	f := &FD{fd: fd, mtu: mtu, LinkEndpoint: ep}

	return f, err
}

func (f *FD) Type() string {
	return Driver
}

func (f *FD) Name() string {
	return strconv.Itoa(f.fd)
}

func (f *FD) Close() error {
	return unix.Close(f.fd)
}

var _ device.Device = (*FD)(nil)

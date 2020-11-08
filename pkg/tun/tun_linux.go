package tun

import (
	"errors"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/link/rawfile"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type linuxTun struct {
	stack.LinkEndpoint

	tunName string
	tunFile *os.File
}

func CreateTUN(name string, n uint32) (Device, error) {
	fd, err := tun.Open(name)
	if err != nil {
		return nil, err
	}

	if n > 0 {
		if err := setMTU(name, n); err != nil {
			return nil, err
		}
	}

	var mtu uint32
	if mtu, err = rawfile.GetMTU(name); err != nil {
		return nil, err
	}

	var ep stack.LinkEndpoint
	if ep, err = fdbased.New(&fdbased.Options{
		FDs: []int{fd},
		MTU: mtu,
		// TUN only
		EthernetHeader: false,
	}); err != nil {
		return nil, err
	}

	return &linuxTun{
		LinkEndpoint: ep,
		tunName:      name,
		tunFile:      os.NewFile(uintptr(fd), "tun"),
	}, nil
}

func (t *linuxTun) Name() string {
	return t.tunName
}

func (t *linuxTun) Close() error {
	return t.tunFile.Close()
}

func setMTU(name string, n uint32) error {
	// open datagram socket
	fd, err := unix.Socket(
		unix.AF_INET,
		unix.SOCK_DGRAM,
		0,
	)

	if err != nil {
		return err
	}

	defer unix.Close(fd)

	const ifReqSize = unix.IFNAMSIZ + 64

	// do ioctl call
	var ifr [ifReqSize]byte
	copy(ifr[:], name)
	*(*uint32)(unsafe.Pointer(&ifr[unix.IFNAMSIZ])) = n
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCSIFMTU),
		uintptr(unsafe.Pointer(&ifr[0])),
	)

	if errno != 0 {
		return errors.New("failed to set MTU of TUN device")
	}

	return nil
}

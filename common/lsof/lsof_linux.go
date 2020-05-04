// +build linux

package lsof

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

type address struct {
	ip      net.IP
	port    uint16
	network string
}

func (a *address) Network() string {
	return a.network
}

func (a *address) String() string {
	return fmt.Sprintf("%s:%d", a.ip.String(), a.port)
}

type socket struct {
	localAddr  address
	remoteAddr address
	inode      int
}

func getSocketList(network string) ([]*socket, error) {
	file := fmt.Sprintf("/proc/net/%s", network)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var socketList []*socket

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// get connection info
		info := make([]string, 0)
		for _, i := range strings.Split(line, " ") {
			if strings.TrimSpace(i) != "" {
				info = append(info, i)
			}
		}
		// length check
		if len(info) < 10 {
			continue
		}

		var localAddr address
		var remoteAddr address
		rawLocalAddr, rawRemoteAddr, rawInode := info[1], info[2], info[9]
		if ip, port, err := parseAddr(rawLocalAddr); err != nil {
			continue
		} else {
			localAddr.network, localAddr.ip, localAddr.port = network, ip, port
		}

		if ip, port, err := parseAddr(rawRemoteAddr); err != nil {
			continue
		} else {
			remoteAddr.network, remoteAddr.ip, remoteAddr.port = network, ip, port
		}

		inode, err := strconv.Atoi(rawInode)
		if err != nil {
			continue
		}

		socketList = append(socketList, &socket{
			localAddr:  localAddr,
			remoteAddr: remoteAddr,
			inode:      inode,
		})
	}

	return socketList, nil
}

func parseIPv4(s string) (net.IP, error) {
	if len(s) != net.IPv4len*2 {
		return nil, fmt.Errorf("bad format: %s", s)
	}
	ip := make(net.IP, net.IPv4len)
	ipNum, _ := strconv.ParseUint(s, 16, 32)
	binary.LittleEndian.PutUint32(ip, uint32(ipNum))
	return ip, nil
}

func parseIPv6(s string) (net.IP, error) {
	if len(s) != net.IPv6len*2 {
		return nil, fmt.Errorf("bad format: %s", s)
	}
	ip := make(net.IP, net.IPv6len)
	for i := 0; i < net.IPv6len/4; i++ {
		ipNum, _ := strconv.ParseUint(s[i*8:(i+1)*8], 16, 32)
		binary.LittleEndian.PutUint32(ip[i*4:(i+1)*4], uint32(ipNum))
	}
	return ip, nil
}

func parseAddr(raw string) (ip net.IP, port uint16, err error) {
	addr := strings.Split(raw, ":")
	if len(addr) != 2 {
		err = fmt.Errorf("IP format error")
		return
	}

	switch len(addr[0]) {
	case net.IPv4len * 2:
		ip, err = parseIPv4(addr[0])
	case net.IPv6len * 2:
		ip, err = parseIPv6(addr[0])
	default:
		err = fmt.Errorf("bad format: %s", addr[0])
	}
	if err != nil {
		return
	}

	portLong, err := strconv.ParseUint(addr[1], 16, 16)
	if err != nil {
		return
	}
	port = uint16(portLong)
	return
}

func getCommandNameByPID(pid int) (string, error) {
	comm := fmt.Sprintf("/proc/%d/comm", pid)
	name, err := ioutil.ReadFile(comm)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(name)), nil
}

func getAllPID() ([]int, error) {
	files, err := ioutil.ReadDir("/proc/")
	if err != nil {
		return nil, err
	}

	var pidList []int
	for _, f := range files {
		if f.IsDir() {
			// Dir name isDigit
			pid, err := strconv.Atoi(f.Name())
			if err != nil {
				continue
			}
			pidList = append(pidList, pid)
		}
	}
	return pidList, nil
}

func getPIDSocketInode(pid int) ([]int, error) {
	dirname := fmt.Sprintf("/proc/%d/fd/", pid)
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	var inodeList []int
	var prefix = "socket:["
	for _, fd := range files {
		name, err := os.Readlink(dirname + fd.Name())
		if err != nil {
			// ignore error, fd may released
			continue
		}
		if strings.HasPrefix(name, prefix) {
			inode, err := strconv.Atoi(name[len(prefix) : len(name)-1])
			if err != nil {
				return nil, fmt.Errorf("unknown format: %s", name)
			}
			inodeList = append(inodeList, inode)
		}
	}
	return inodeList, nil
}

func GetCommandNameBySocket(network string, addr string, port uint16) (comm string, err error) {
	socketList, err := getSocketList(network)
	if err != nil {
		return
	}

	var inode = -1 // negative init
	patten := fmt.Sprintf("%s:%d", addr, port)
	for _, socket := range socketList {
		if patten == socket.localAddr.String() {
			inode = socket.inode
			break // best match
		} else if port == socket.localAddr.port && socket.localAddr.ip.IsUnspecified() {
			inode = socket.inode // for udp compatible
		}
	}

	if inode == -1 {
		return "", ErrNotFound
	}

	pidList, err := getAllPID()
	if err != nil {
		err = fmt.Errorf("get all PID failed: %v", err)
		return
	}

	for _, pid := range pidList {
		inodeList, err := getPIDSocketInode(pid)
		if err != nil {
			// ignore error
			continue
		}
		for _, i := range inodeList {
			if i == inode {
				name, err := getCommandNameByPID(pid)
				if err != nil {
					// ignore error
					continue
				}
				return name, nil
			}
		}
	}
	return "", ErrNotFound
}

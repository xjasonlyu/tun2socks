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

type addr struct {
	ip net.IP
	port uint16
	network string
}

func (a *addr) Network() string {
	return a.network
}

func (a *addr) String() string {
	return fmt.Sprintf("%s:%d", a.ip.String(), a.port)
}

type socket struct {
	localAddr addr
	remoteAddr addr
	inode int
}

func getSocketTable(network string) ([]*socket, error) {
	file := fmt.Sprintf("/proc/net/%s", network)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var sockets []*socket

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// get connection info
		info := make([]string, 0)
		for _, i := range strings.Split(line, " ") {
			if strings.TrimSpace(i) != "" {
				info = append(info ,i)
			}
		}
		// length check
		if len(info) < 10 {
			continue
		}
		var localAddr addr
		var remoteAddr addr
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

		sockets = append(sockets, &socket{
			localAddr: localAddr,
			remoteAddr: remoteAddr,
			inode: inode,
		})
	}

	return sockets, nil
}

// IPv4 Only
func parseAddr(raw string) (ip net.IP, port uint16, err error) {
	addr := strings.Split(raw, ":")
	if len(addr) != 2 {
		err = fmt.Errorf("IP format error")
		return
	}

	ipLong, err := strconv.ParseUint(addr[0], 16, 32)
	if err != nil {
		return
	}
	ip = make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, uint32(ipLong))

	portLong, err := strconv.ParseUint(addr[1], 16, 16)
	if err != nil {
		return
	}
	port = uint16(portLong)
	return
}

func getCommandNameByPID(pid int) (string, error) {
	file := fmt.Sprintf("/proc/%d/comm", pid)
	name, err := ioutil.ReadFile(file)
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
	for _, f := range files {
		name, err := os.Readlink(dirname +f.Name())
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(name, "socket:") {
			name = strings.TrimPrefix(name, "socket:[")
			name = strings.TrimSuffix(name, "]")
			inode, err := strconv.Atoi(name)
			if err != nil {
				return nil, fmt.Errorf("unknown format: %s", name)
			}
			inodeList = append(inodeList, inode)
		}
	}
	return inodeList, nil
}

func GetCommandNameBySocket(network string, addr string, port uint16) (comm string, err error) {
	socketTable, err := getSocketTable(network)
	if err != nil {
		return
	}

	var inode int
	patten := fmt.Sprintf("%s:%d", addr, port)
	for _, socket := range socketTable {
		if patten == socket.localAddr.String() {
			inode = socket.inode
			break
		}
	}

	if inode == 0 {
		return "", ErrNotFound
	}

	pidList, err := getAllPID()
	if err != nil {
		err = fmt.Errorf("get all PID error: %v", err)
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

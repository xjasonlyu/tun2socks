// +build darwin,!ios linux,!android

package lsof

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func GetCommandNameBySocket(network string, addr string, port uint16) (string, error) {
	pattern := ""
	switch network {
	case "tcp":
		pattern = fmt.Sprintf("-i%s@%s:%d", network, addr, port)
	case "udp":
		// The current approach isn't quite accurate for
		// udp sockets, as more than one processes can
		// listen on the same udp port. Moreover, if
		// the application closes the socket immediately
		// after sending out the packet (e.g. it just
		// uploading data but not receving any data),
		// we may not be able to find it.
		pattern = fmt.Sprintf("-i%s:%d", network, port)
	default:
	}
	out, err := exec.Command("lsof", "-n", "-Fc", pattern).Output()
	if err != nil {
		if len(out) != 0 {
			return "", errors.New(fmt.Sprintf("%v, output: %s", err, out))
		}
		return "", err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		// There may be multiple candidate
		// sockets in the list, just take
		// the first one for simplicity.
		if strings.HasPrefix(line, "c") {
			return line[1:len(line)], nil
		}
	}
	return "", errors.New("not found")
}

package metadata

import (
	"fmt"
)

const (
	TCP Network = iota
	UDP
)

type Network uint8

func (n Network) String() string {
	switch n {
	case TCP:
		return "tcp"
	case UDP:
		return "udp"
	default:
		return fmt.Sprintf("network(%d)", n)
	}
}

func (n Network) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

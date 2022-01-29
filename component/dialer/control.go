package dialer

import (
	"errors"
	"net"
	"syscall"
)

type controlFunc func(string, string, syscall.RawConn) error

var _controlPool = make([]controlFunc, 0, 2)

func addControl(f controlFunc) {
	_controlPool = append(_controlPool, f)
}

func setControl(i interface{}) {
	control := func(address, network string, c syscall.RawConn) error {
		for _, f := range _controlPool {
			if err := f(address, network, c); err != nil {
				return err
			}
		}
		return nil
	}

	switch v := i.(type) {
	case *net.Dialer:
		v.Control = control
	case *net.ListenConfig:
		v.Control = control
	default:
		panic(errors.New("wrong type"))
	}
}

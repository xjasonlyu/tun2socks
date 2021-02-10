package dialer

import (
	"net"
	"sync"
)

var _bindOnce sync.Once

// BindToInterface binds dialer to specific interface.
func BindToInterface(name string) error {
	i, err := net.InterfaceByName(name)
	if err != nil {
		return err
	}

	_bindOnce.Do(func() {
		addControl(bindToInterface(i))
	})
	return nil
}

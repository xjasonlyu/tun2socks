package dialer

import "net"

var _boundInterface *net.Interface

// BindToInterface binds dialer to specific interface.
func BindToInterface(name string) error {
	i, err := net.InterfaceByName(name)
	if err != nil {
		return err
	}
	_boundInterface = i
	return nil
}

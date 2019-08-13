// +build ios android

package lsof

func GetCommandNameBySocket(network string, addr string, port uint16) (string, error) {
	return "", ErrNotImplemented
}

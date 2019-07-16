// +build ios android

package lsof

import (
	"errors"
)

func GetCommandNameBySocket(network string, addr string, port uint16) (string, error) {
	return "", errors.New("not implemented")
}

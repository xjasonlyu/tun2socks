package fdbased

import (
	"errors"

	"github.com/TianHe-Labs/Zeus/core/device"
)

func Open(name string, mtu uint32) (device.Device, error) {
	return nil, errors.New("not supported")
}

package dialer

import (
	"sync"
)

var _setOnce sync.Once

// SetMark sets the mark for each packet sent through this dialer(socket).
func SetMark(i int) {
	_setOnce.Do(func() {
		addControl(setMark(i))
	})
}

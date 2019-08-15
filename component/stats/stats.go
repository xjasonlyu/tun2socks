package stats

import (
	C "github.com/xjasonlyu/tun2socks/constant"
)

type SessionStater interface {
	Start() error
	Stop() error
	AddSession(key interface{}, session *C.Session)
	GetSession(key interface{}) *C.Session
	RemoveSession(key interface{})
}

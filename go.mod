module github.com/xjasonlyu/tun2socks

go 1.16

require (
	github.com/Dreamacro/go-shadowsocks2 v0.1.7
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/go-chi/render v1.0.1
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/google/btree v1.0.1 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/atomic v1.7.0
	golang.org/x/net v0.0.0-20210316092652-d523dce5a7f4 // indirect
	golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.zx2c4.com/wireguard v0.0.0-20210225140808-70b7b7158fc9
	gvisor.dev/gvisor v0.0.0-20210318191957-7fac7e32f3a8
)

replace gvisor.dev/gvisor => github.com/xjasonlyu/gvisor v0.0.0-20210321122453-eb40de9b30e3

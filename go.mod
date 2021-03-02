module github.com/xjasonlyu/tun2socks

go 1.16

require (
	github.com/Dreamacro/go-shadowsocks2 v0.1.6
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/go-chi/render v1.0.1
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/magefile/mage v1.11.0 // indirect
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/atomic v1.7.0
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sys v0.0.0-20210301091718-77cc2087c03b
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.zx2c4.com/wireguard v0.0.0-20210225140808-70b7b7158fc9
	gvisor.dev/gvisor v0.0.0-20210301201720-865ca64ee8c0
)

replace gvisor.dev/gvisor => github.com/xjasonlyu/gvisor v0.0.0-20210302152121-27a6751e19c2

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
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/net v0.0.0-20210220033124-5f55cee0dc0d // indirect
	golang.org/x/sys v0.0.0-20210220050731-9a76102bfb43
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.zx2c4.com/wireguard v0.0.0-20210222013812-386a61306bd6
	gvisor.dev/gvisor v0.0.0-20210220013851-93fc09248a2f
)

replace gvisor.dev/gvisor => github.com/xjasonlyu/gvisor v0.0.0-20210222043145-15b944d12b4f

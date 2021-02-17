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
	github.com/sirupsen/logrus v1.7.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/atomic v1.7.0
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/sys v0.0.0-20210216224549-f992740a1bac
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324
	golang.zx2c4.com/wireguard v0.0.0-20210216200525-4e439ea10e32
	gvisor.dev/gvisor v0.0.0-20210213011838-e7ae604b523e
)

replace gvisor.dev/gvisor => github.com/xjasonlyu/gvisor v0.0.0-20210217091018-5497a9345bcd

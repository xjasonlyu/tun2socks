package tun

type Option func(*TUN)

func WithName(name string) Option {
	return func(t *TUN) {
		t.name = name
	}
}

func WithMTU(mtu uint32) Option {
	return func(t *TUN) {
		t.mtu = mtu
	}
}

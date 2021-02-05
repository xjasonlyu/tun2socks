package tun

func WithComponentID(componentID string) Option {
	return func(t *TUN) {
		t.componentID = componentID
	}
}

func WithNetwork(network string) Option {
	return func(t *TUN) {
		t.network = network
	}
}

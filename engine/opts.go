package engine

type Option func(*Engine)

func WithDevice(device string) Option {
	return func(e *Engine) {
		e.rawDevice = device
	}
}

func WithInterface(iface string) Option {
	return func(e *Engine) {
		e.iface = iface
	}
}

func WithLogLevel(level string) Option {
	return func(e *Engine) {
		e.logLevel = level
	}
}

func WithMTU(mtu int) Option {
	return func(e *Engine) {
		e.mtu = uint32(mtu)
	}
}

func WithProxy(proxy string) Option {
	return func(e *Engine) {
		e.rawProxy = proxy
	}
}

func WithStats(stats, secret string) Option {
	return func(e *Engine) {
		e.stats = stats
		e.secret = secret
	}
}

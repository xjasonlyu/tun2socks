package core

import (
	"context"
	"errors"
	"time"
)

// DeferredTCPConfig enables remote-first TCP when supplied in Config.DeferredTCP.
// Zero cache values select defaults; negative values are invalid. Hits never
// refresh the expiry. Configuration is copied when the stack is created.
type DeferredTCPConfig struct {
	// Cancel Context before closing the stack to release pending dials, active
	// TCP sockets and cached failures. Nil uses context.Background().
	Context context.Context
	// FailureCacheTTL starts at dial failure; zero defaults to 30 seconds.
	FailureCacheTTL time.Duration
	// FailureCacheCapacity is shared by IPv4/IPv6; zero defaults to 4096.
	FailureCacheCapacity int
}

func (c DeferredTCPConfig) normalized() (DeferredTCPConfig, error) {
	if c.FailureCacheTTL < 0 || c.FailureCacheCapacity < 0 {
		return c, errors.New("TCP failure cache TTL and capacity must not be negative")
	}
	if c.Context == nil {
		c.Context = context.Background()
	}
	if c.FailureCacheTTL == 0 {
		c.FailureCacheTTL = tcpFailureTTL
	}
	if c.FailureCacheCapacity == 0 {
		c.FailureCacheCapacity = tcpFailureCapacity
	}
	return c, nil
}

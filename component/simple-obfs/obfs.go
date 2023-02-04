// Package obfs provides obfuscation functionality for Shadowsocks protocol.
package obfs

import (
	"crypto/rand"
	"math"
	"math/big"
)

// Ref: github.com/Dreamacro/clash/component/simple-obfs

func randInt() int {
	n, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt))
	return int(n.Int64())
}

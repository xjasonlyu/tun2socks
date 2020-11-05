package tunnel

import (
	"fmt"

	"github.com/xjasonlyu/clash/component/resolver"
	"github.com/xjasonlyu/tun2socks/internal/adapter"
)

func generateNATKey(m *adapter.Metadata) string {
	return m.SourceAddress() /* Full Cone NAT Key */
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func resolveMetadata(metadata *adapter.Metadata) error {
	if metadata.DstIP == nil {
		return fmt.Errorf("destination IP is nil")
	}

	if resolver.IsFakeIP(metadata.DstIP) {
		var exist bool
		metadata.Host, exist = resolver.FindHostByIP(metadata.DstIP)
		if !exist {
			return fmt.Errorf("fake DNS record %s missing", metadata.DstIP)
		}
	}

	return nil
}

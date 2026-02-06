package dialer

func isTCPSocket(network string) bool {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return true
	default:
		return false
	}
}

func isUDPSocket(network string) bool {
	switch network {
	case "udp", "udp4", "udp6":
		return true
	default:
		return false
	}
}

func isICMPSocket(network string) bool {
	switch network {
	case "ip:icmp", "ip4:icmp", "ip6:ipv6-icmp":
		return true
	case "ip4", "ip6":
		return true
	default:
		return false
	}
}

package dns

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/xjasonlyu/clash/component/fakeip"
	"github.com/xjasonlyu/clash/component/trie"
	"github.com/xjasonlyu/clash/dns"
	"github.com/xjasonlyu/tun2socks/pkg/log"
)

func parseFakeIP(fakeIPRange string, fakeIPFilter []string) (*fakeip.Pool, error) {
	_, ipnet, err := net.ParseCIDR(fakeIPRange)
	if err != nil {
		return nil, err
	}

	var host *trie.DomainTrie
	// fake ip skip host filter
	if len(fakeIPFilter) != 0 {
		host = trie.New()
		for _, domain := range fakeIPFilter {
			_ = host.Insert(domain, true)
		}
	}

	return fakeip.New(ipnet, defaultCacheSize, host)
}

func parseHosts(hosts map[string]string) (*trie.DomainTrie, error) {
	tree := trie.New()

	// add default hosts
	if err := tree.Insert("localhost", net.IP{127, 0, 0, 1}); err != nil {
		log.Errorf("[DNS] insert localhost to host error: %v", err)
	}

	if len(hosts) != 0 {
		for domain, ipStr := range hosts {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return nil, fmt.Errorf("%s is not a valid IP", ipStr)
			}
			_ = tree.Insert(domain, ip)
		}
	}

	return tree, nil
}

func parseHostsSlice(s []string) map[string]string {
	m := make(map[string]string)

	for _, i := range s {
		if strings.Contains(i, "=") {
			v := strings.SplitN(i, "=", 2)
			m[v[0]] = v[1]
			continue
		}
		if strings.Contains(i, ":") {
			v := strings.SplitN(i, ":", 2)
			m[v[0]] = v[1]
			continue
		}
		log.Debugf("invalid hosts item: %s", i)
	}

	return m
}

func hostWithDefaultPort(host string, defPort string) (string, error) {
	if !strings.Contains(host, ":") {
		host += ":"
	}

	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		return "", err
	}

	if port == "" {
		port = defPort
	}

	return net.JoinHostPort(hostname, port), nil
}

func parseNameServer(servers []string) ([]dns.NameServer, error) {
	var nameservers []dns.NameServer

	for idx, server := range servers {
		server := strings.TrimSpace(server)
		if server == "" {
			return nil, fmt.Errorf("empty server field")
		}

		// parse without scheme .e.g 8.8.8.8:53
		if !strings.Contains(server, "://") {
			server = "udp://" + server
		}
		u, err := url.Parse(server)
		if err != nil {
			return nil, fmt.Errorf("DNS NameServer[%d] format: %s", idx, err.Error())
		}

		var addr, dnsNetType string
		switch u.Scheme {
		case "udp":
			addr, err = hostWithDefaultPort(u.Host, "53")
			dnsNetType = "" // UDP
		case "tcp":
			addr, err = hostWithDefaultPort(u.Host, "53")
			dnsNetType = "tcp" // TCP
		case "tls":
			addr, err = hostWithDefaultPort(u.Host, "853")
			dnsNetType = "tcp-tls" // DNS over TLS
		case "https":
			clearURL := url.URL{Scheme: "https", Host: u.Host, Path: u.Path}
			addr = clearURL.String()
			dnsNetType = "https" // DNS over HTTPS
		default:
			return nil, fmt.Errorf("DNS NameServer[%d] unsupport scheme: %s", idx, u.Scheme)
		}

		if err != nil {
			return nil, fmt.Errorf("DNS NameServer[%d] format: %s", idx, err.Error())
		}

		nameservers = append(
			nameservers,
			dns.NameServer{
				Net:  dnsNetType,
				Addr: addr,
			},
		)
	}
	return nameservers, nil
}

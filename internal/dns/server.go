package dns

import (
	"errors"
	"net/url"
	"strings"

	"github.com/xjasonlyu/clash/component/resolver"
	"github.com/xjasonlyu/clash/dns"
)

const (
	defaultScheme      = "dns"
	defaultCacheSize   = 1000
	defaultFakeIPRange = "198.18.0.0/15"
)

var _defaultNameServer = []string{"223.5.5.5", "8.8.8.8"}

func Start(dnsURL string, rawHosts []string) error {
	if !strings.Contains(dnsURL, "://") {
		dnsURL = defaultScheme + "://" + dnsURL
	}

	u, err := url.Parse(dnsURL)
	if err != nil {
		return err
	}

	if strings.ToLower(u.Scheme) != defaultScheme {
		return errors.New("unsupported scheme")
	}

	serverAddr := u.Host
	if serverAddr == "" {
		return nil
	}

	fakeIPRange := u.Query().Get("fake-ip-range")
	if fakeIPRange == "" {
		fakeIPRange = defaultFakeIPRange
	}

	var fakeIPFilter []string
	if raw := strings.TrimSpace(u.Query().Get("fake-ip-filter")); raw != "" {
		fakeIPFilter = append(fakeIPFilter, strings.Split(raw, ",")...)
	}

	pool, err := parseFakeIP(fakeIPRange, fakeIPFilter)
	if err != nil {
		return err
	}

	var mainNS, defaultNS []dns.NameServer
	{
		var ns []string
		if raw := strings.TrimSpace(u.Query().Get("default-nameserver")); raw != "" {
			ns = append(ns, strings.Split(raw, ",")...)
		}

		if len(ns) == 0 {
			ns = append(ns, _defaultNameServer...)
		}

		if defaultNS, err = parseNameServer(ns); err != nil {
			return err
		}
	}
	{
		var ns []string
		if raw := strings.TrimSpace(u.Query().Get("nameserver")); raw != "" {
			ns = append(ns, strings.Split(raw, ",")...)
		}

		if len(ns) == 0 {
			ns = append(ns, _defaultNameServer...)
		}

		if mainNS, err = parseNameServer(ns); err != nil {
			return err
		}
	}

	hosts, err := parseHosts(parseHostsSlice(rawHosts))
	if err != nil {
		return err
	}

	cfg := dns.Config{
		IPv6:         true,
		Pool:         pool,
		Hosts:        hosts,
		Main:         mainNS,
		Default:      defaultNS,
		EnhancedMode: dns.FAKEIP,
	}

	r := dns.NewResolver(cfg)
	m := dns.NewEnhancer(cfg)

	// reuse cache of old host mapper
	if old := resolver.DefaultHostMapper; old != nil {
		m.PatchFrom(old.(*dns.ResolverEnhancer))
	}

	resolver.DefaultResolver = r
	resolver.DefaultHostMapper = m

	return dns.ReCreateServer(serverAddr, r, m)
}

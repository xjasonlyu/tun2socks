package proxy

import (
	"errors"
	"net/url"
	"strings"
)

// URL is the universal representation of the proxy configuration.
type URL url.URL

func (u *URL) Protocol() string {
	return u.Scheme
}

func (u *URL) Address() string {
	return u.Host
}

func (u *URL) String() string {
	return (&url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   strings.TrimRight(u.Path, "/"),
	}).String()
}

func ParseURL(rawURL string) (*URL, error) {
	proxyURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if proxyURL.Scheme == "" {
		return nil, errors.New("proxy: protocol not specified")
	}
	return (*URL)(proxyURL), nil
}

func MustParseURL(rawURL string) *URL {
	u, err := ParseURL(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

package http

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/utils"
)

var _ proxy.Proxy = (*HTTP)(nil)

type HTTP struct {
	addr string
	user string
	pass string
}

func New(addr, user, pass string) (*HTTP, error) {
	return &HTTP{
		addr: addr,
		user: user,
		pass: pass,
	}, nil
}

func (h *HTTP) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	c, err = dialer.DialContext(ctx, "tcp", h.addr)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", h.addr, err)
	}
	utils.SetKeepAlive(c)

	defer func(c net.Conn) {
		utils.SafeConnClose(c, err)
	}(c)

	err = h.shakeHand(metadata, c)
	return c, err
}

func (h *HTTP) DialUDP(*M.Metadata) (net.PacketConn, error) {
	return nil, errors.ErrUnsupported
}

func (h *HTTP) shakeHand(metadata *M.Metadata, rw io.ReadWriter) error {
	addr := metadata.DestinationAddress()
	req := &http.Request{
		Method: http.MethodConnect,
		URL: &url.URL{
			Host: addr,
		},
		Host: addr,
		Header: http.Header{
			"Proxy-Connection": []string{"Keep-Alive"},
		},
	}

	if h.user != "" && h.pass != "" {
		req.Header.Set("Proxy-Authorization", fmt.Sprintf("Basic %s", basicAuth(h.user, h.pass)))
	}

	if err := req.Write(rw); err != nil {
		return err
	}

	resp, err := http.ReadResponse(bufio.NewReader(rw), req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusProxyAuthRequired:
		return errors.New("HTTP auth required by proxy")
	case http.StatusMethodNotAllowed:
		return errors.New("CONNECT method not allowed by proxy")
	default:
		return fmt.Errorf("HTTP connect status: %s", resp.Status)
	}
}

// The Basic authentication scheme is based on the model that the client
// needs to authenticate itself with a user-id and a password for each
// protection space ("realm"). The realm value is a free-form string
// that can only be compared for equality with other realms on that
// server. The server will service the request only if it can validate
// the user-id and password for the protection space applying to the
// requested resource.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func Parse(u *url.URL) (proxy.Proxy, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()
	return New(address, username, password)
}

func init() {
	proxy.RegisterProtocol("http", Parse)
}

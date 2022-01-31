package proxy

// Ref: https://github.com/Dreamacro/clash/adapter/outbound/http

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

	"github.com/xjasonlyu/tun2socks/v2/component/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy/proto"
)

type HTTP struct {
	*Base

	user string
	pass string
}

func NewHTTP(addr, user, pass string) (*HTTP, error) {
	return &HTTP{
		Base: &Base{
			addr:  addr,
			proto: proto.HTTP,
		},
		user: user,
		pass: pass,
	}, nil
}

func (h *HTTP) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	c, err = dialer.DialContext(ctx, "tcp", h.Addr())
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", h.Addr(), err)
	}
	setKeepAlive(c)

	defer safeConnClose(c, err)

	err = h.shakeHand(metadata, c)
	return
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
		auth := h.user + ":" + h.pass
		req.Header.Add("Proxy-Authorization",
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth))))
	}

	if err := req.Write(rw); err != nil {
		return err
	}

	resp, err := http.ReadResponse(bufio.NewReader(rw), req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusProxyAuthRequired {
		return errors.New("HTTP need auth")
	}

	if resp.StatusCode == http.StatusMethodNotAllowed {
		return errors.New("CONNECT method not allowed by proxy")
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return errors.New(resp.Status)
	}

	return fmt.Errorf("HTTP connect status code: %d", resp.StatusCode)
}

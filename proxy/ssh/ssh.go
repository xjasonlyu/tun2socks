package ssh

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/utils"
)

type SSH struct {
	addr   string
	config *ssh.ClientConfig
}

func New(addr, user, pass, keyFile, passphrase string) (*SSH, error) {
	var auth []ssh.AuthMethod
	if pass != "" {
		auth = append(auth, ssh.Password(pass))
	}
	if keyFile != "" {
		key, err := os.ReadFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("ssh: read file: %w", err)
		}
		var signer ssh.Signer
		if passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(
				key, []byte(passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, fmt.Errorf("ssh: parse private key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	return &SSH{
		addr: addr,
		config: &ssh.ClientConfig{
			User:            user,
			Auth:            auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         utils.TCPConnectTimeout,
		},
	}, nil
}

func (s *SSH) DialContext(ctx context.Context, metadata *M.Metadata) (_ net.Conn, err error) {
	c, err := dialer.DialContext(ctx, "tcp", s.addr)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", s.addr, err)
	}
	utils.SetKeepAlive(c)

	defer func(c net.Conn) {
		utils.SafeConnClose(c, err)
	}(c)

	sc, ch, reqs, err := ssh.NewClientConn(c, s.addr, s.config)
	if err != nil {
		return nil, err
	}

	client := ssh.NewClient(sc, ch, reqs)
	conn, err := client.Dial("tcp", metadata.DestinationAddress())
	if err != nil {
		client.Close()
		return nil, err
	}
	return &sshConn{conn, client}, nil
}

func (s *SSH) DialUDP(*M.Metadata) (net.PacketConn, error) {
	return nil, errors.ErrUnsupported
}

type sshConn struct {
	net.Conn
	client *ssh.Client
}

func (c *sshConn) Close() error {
	defer c.client.Close()
	return c.Conn.Close()
}

func Parse(u *url.URL) (proxy.Proxy, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()
	keyFile := u.Query().Get("privateKeyFile")
	passphrase := u.Query().Get("passphrase")
	return New(address, username, password, keyFile, passphrase)
}

func init() {
	proxy.RegisterProtocol("ssh", Parse)
}

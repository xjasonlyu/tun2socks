package masque_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/quic-go/masque-go"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/yosida95/uritemplate/v3"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	masquecli "github.com/xjasonlyu/tun2socks/v2/proxy/masque"
)

// generateCert creates a self-signed ECDSA cert for 127.0.0.1 / localhost
// and writes PEM files into dir.
func generateCert(t *testing.T, dir string) (certFile, keyFile string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	tpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "masque-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:              []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}

// startEcho opens a loopback UDP echo server.
func startEcho(t *testing.T) *net.UDPAddr {
	t.Helper()
	c, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	go func() {
		buf := make([]byte, 2048)
		for {
			n, addr, err := c.ReadFromUDP(buf)
			if err != nil {
				return
			}
			_, _ = c.WriteToUDP(buf[:n], addr)
		}
	}()
	return c.LocalAddr().(*net.UDPAddr)
}

type proxyOpts struct {
	requireAuth string // if non-empty, require this Proxy-Authorization value
	failAll     bool   // return 403 to everything
}

// startProxy starts a MASQUE server on localhost and returns its address.
func startProxy(t *testing.T, opts proxyOpts) *net.UDPAddr {
	t.Helper()
	dir := t.TempDir()
	certFile, keyFile := generateCert(t, dir)

	lc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	addr := lc.LocalAddr().(*net.UDPAddr)

	tplRaw := fmt.Sprintf("https://127.0.0.1:%d/.well-known/masque/udp/{target_host}/{target_port}/", addr.Port)
	tpl := uritemplate.MustNew(tplRaw)
	mp := &masque.Proxy{}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/masque/udp/", func(w http.ResponseWriter, r *http.Request) {
		if opts.failAll {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if opts.requireAuth != "" {
			got := r.Header.Get("Proxy-Authorization")
			if got != opts.requireAuth {
				http.Error(w, "auth required", http.StatusProxyAuthRequired)
				return
			}
		}
		mreq, err := masque.ParseRequest(r, tpl)
		if err != nil {
			var perr *masque.RequestParseError
			if errors.As(err, &perr) {
				w.WriteHeader(perr.HTTPStatus)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := mp.Proxy(w, mreq); err != nil {
			t.Logf("proxy.Proxy: %v", err)
		}
	})

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	server := &http3.Server{
		Handler:         mux,
		TLSConfig:       &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{http3.NextProtoH3}},
		EnableDatagrams: true,
		QUICConfig:      &quic.Config{EnableDatagrams: true},
	}
	errCh := make(chan error, 1)
	go func() { errCh <- server.Serve(lc) }()

	// Give Serve a moment to perform its synchronous setDF check. On
	// sandboxes that prohibit setting the DF bit (e.g. network-restricted
	// CI containers) it fails immediately and we skip the test.
	select {
	case err := <-errCh:
		if err != nil && strings.Contains(err.Error(), "setting DF") {
			_ = lc.Close()
			t.Skipf("skipping: environment does not support setting UDP DF bit: %v", err)
		}
		t.Fatalf("http3.Server.Serve exited early: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	t.Cleanup(func() {
		_ = server.Close()
		_ = lc.Close()
		_ = mp.Close()
		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Logf("http3.Server.Serve: %v", err)
		}
	})
	return addr
}

func basicAuth(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func echoMetadata(echo *net.UDPAddr) *M.Metadata {
	return &M.Metadata{
		Network: M.UDP,
		DstIP:   netip.AddrFrom4([4]byte{127, 0, 0, 1}),
		DstPort: uint16(echo.Port),
	}
}

func parseProxy(t *testing.T, raw string) *masquecli.Masque {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	p, err := masquecli.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	m := p.(*masquecli.Masque)
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func TestMasqueEchoLoopback(t *testing.T) {
	paddr := startProxy(t, proxyOpts{})
	echo := startEcho(t)

	m := parseProxy(t, fmt.Sprintf("masque://%s/?insecure=true&sni=localhost", paddr.String()))
	pc, err := m.DialUDP(echoMetadata(echo))
	if err != nil {
		t.Fatalf("DialUDP: %v", err)
	}
	defer pc.Close()

	want := []byte("hello masque")
	if _, err := pc.WriteTo(want, nil); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	_ = pc.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 2048)
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	if !bytes.Equal(buf[:n], want) {
		t.Fatalf("echo mismatch: got %q want %q", buf[:n], want)
	}
}

func TestMasqueBasicAuthRequired(t *testing.T) {
	paddr := startProxy(t, proxyOpts{requireAuth: basicAuth("alice", "s3cr3t")})
	echo := startEcho(t)

	// Without credentials: must fail with 407.
	m1 := parseProxy(t, fmt.Sprintf("masque://%s/?insecure=true&sni=localhost", paddr.String()))
	if _, err := m1.DialUDP(echoMetadata(echo)); err == nil {
		t.Fatal("expected auth failure, got nil")
	} else if !strings.Contains(err.Error(), "407") {
		t.Fatalf("expected 407 in error, got: %v", err)
	}

	// With credentials: must succeed and echo.
	m2 := parseProxy(t, fmt.Sprintf("masque://alice:s3cr3t@%s/?insecure=true&sni=localhost", paddr.String()))
	pc, err := m2.DialUDP(echoMetadata(echo))
	if err != nil {
		t.Fatalf("DialUDP with creds: %v", err)
	}
	defer pc.Close()
	if _, err := pc.WriteTo([]byte("ping"), nil); err != nil {
		t.Fatal(err)
	}
	_ = pc.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 2048)
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	if !bytes.Equal(buf[:n], []byte("ping")) {
		t.Fatalf("echo mismatch: got %q", buf[:n])
	}
}

func TestMasqueRejectsNon2xx(t *testing.T) {
	paddr := startProxy(t, proxyOpts{failAll: true})
	echo := startEcho(t)

	m := parseProxy(t, fmt.Sprintf("masque://%s/?insecure=true&sni=localhost", paddr.String()))
	_, err := m.DialUDP(echoMetadata(echo))
	if err == nil {
		t.Fatal("expected non-2xx error, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected 403 in error, got: %v", err)
	}
}

func TestMasqueParseURL(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"plain", "masque://proxy.example.com:443", false},
		{"with-auth", "masque://user:pass@proxy.example.com:443", false},
		{"default-port", "masque://proxy.example.com", false},
		{"custom-template", "masque://proxy.example.com:443/custom/{target_host}/{target_port}/", false},
		{"query-options", "masque://proxy.example.com:443/?sni=foo&insecure=true&alpn=h3,h3-29", false},
		{"template-missing-target-host", "masque://proxy.example.com:443/?template=https://proxy.example.com/x/{target_port}", true},
		{"empty-host", "masque://", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u, err := url.Parse(c.raw)
			if err != nil {
				t.Fatalf("url.Parse: %v", err)
			}
			_, err = masquecli.Parse(u)
			if (err != nil) != c.wantErr {
				t.Fatalf("Parse(%q) err=%v, wantErr=%v", c.raw, err, c.wantErr)
			}
		})
	}
}

func TestMasqueRefusesTCP(t *testing.T) {
	u, _ := url.Parse("masque://127.0.0.1:4433/?insecure=true")
	p, err := masquecli.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	md := &M.Metadata{
		Network: M.TCP,
		DstIP:   netip.AddrFrom4([4]byte{127, 0, 0, 1}),
		DstPort: 80,
	}
	if _, err := p.DialContext(t.Context(), md); err == nil {
		t.Fatal("expected DialContext to refuse, got nil")
	}
}

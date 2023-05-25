package tunnel

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/textproto"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/TianHe-Labs/Zeus/common/pool"
	"github.com/TianHe-Labs/Zeus/core/adapter"
	"github.com/TianHe-Labs/Zeus/log"
	M "github.com/TianHe-Labs/Zeus/metadata"
	"github.com/TianHe-Labs/Zeus/proxy"
	"github.com/TianHe-Labs/Zeus/tunnel/statistic"
)

const (
	// tcpWaitTimeout implements a TCP half-close timeout.
	tcpWaitTimeout = 60 * time.Second
)

func handleTCPConn(originConn adapter.TCPConn) {
	defer originConn.Close()

	id := originConn.ID()
	metadata := &M.Metadata{
		Network: M.TCP,
		SrcIP:   net.IP(id.RemoteAddress.AsSlice()),
		SrcPort: id.RemotePort,
		DstIP:   net.IP(id.LocalAddress.AsSlice()),
		DstPort: id.LocalPort,
	}

	remoteConn, err := proxy.Dial(metadata)
	if err != nil {
		log.Warnf("[TCP] dial %s: %v", metadata.DestinationAddress(), err)
		return
	}
	metadata.MidIP, metadata.MidPort = parseAddr(remoteConn.LocalAddr())

	remoteConn = statistic.DefaultTcpTracker(remoteConn, metadata)

	defer remoteConn.Close()

	log.Infof("[TCP] %s <-> %s", metadata.SourceAddress(), metadata.DestinationAddress())

	pipe(originConn, remoteConn, metadata.DstPort == 80 || metadata.DstPort == 443)
}

// pipe copies copy data to & from provided net.Conn(s) bidirectionally.
func pipe(origin, remote net.Conn, isHTTPorHTTPS bool) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go unidirectionalStream(remote, origin, "origin->remote", &wg, isHTTPorHTTPS)
	go unidirectionalStream(origin, remote, "remote->origin", &wg, false)

	wg.Wait()
}

func unidirectionalStream(dst, src net.Conn, dir string, wg *sync.WaitGroup, needHandled bool) {
	defer wg.Done()
	buf := pool.Get(pool.RelayBufferSize)

	// No need to handle remote->origin
	if needHandled {
		if conn, ok := dst.(*statistic.TcpTracker); ok {
			if !conn.IsHandled {
				// Creating a buffer to keep the data.
				dataBuf := bytes.NewBuffer(make([]byte, 0, len(buf)))
				tee := io.TeeReader(src, dataBuf)

				// Copying data from source to destination
				if _, err := io.CopyBuffer(dst, tee, buf); err != nil {
					log.Debugf("[TCP] copy data for %s: %v", dir, err)
				}

				// Parse domain from this connection
				var domain string
				if conn.Metadata.DstPort == 80 {
					domain = getDomainFromHttpHeader(dataBuf)
				} else if conn.Metadata.DstPort == 443 {
					domain = getDomainFromSNI(dataBuf.Bytes())
				}

				if domain != "" {
					// Mark this connection is handled and save the domain
					log.Infof("[Conn:%s] Host: %s", conn.UUID.String(), domain)
					conn.Metadata.Domain = domain
				}
			}
		}
	} else {
		if _, err := io.CopyBuffer(dst, src, buf); err != nil {
			log.Debugf("[TCP] copy data for %s: %v", dir, err)
		}
	}

	pool.Put(buf)

	// Do the upload/download side TCP half-close.
	if cr, ok := src.(interface{ CloseRead() error }); ok {
		cr.CloseRead()
	}
	if cw, ok := dst.(interface{ CloseWrite() error }); ok {
		cw.CloseWrite()
	}
	// Set TCP half-close timeout.
	dst.SetReadDeadline(time.Now().Add(tcpWaitTimeout))
}

var regex = regexp.MustCompile(`^(?:[a-z0-9-]+\.)+[a-z]+$`)

func isValidDomainName(s string) bool {
	// Simple check: the string is a valid domain if it has at least one dot, and it doesn't start or end with a dot.
	if strings.HasPrefix(s, ".") || strings.HasSuffix(s, ".") {
		return false
	}
	return strings.ContainsRune(s, '.')
}

func getDomainFromSNI(buf []byte) string {
	var sni string
	var prev byte
	for b := 0; b < len(buf)-2; b++ {
		if prev == 0 && buf[b] == 0 {
			start := b + 2
			end := start + int(buf[b+1])
			if start < end && end < len(buf) {
				str := string(buf[start:end])
				// Use a simpler check to see if the string could be a domain before running the regex.
				if isValidDomainName(str) && regex.MatchString(str) {
					sni = str
					continue
				}
			}
		}
		prev = buf[b]
	}
	return sni
}

func getDomainFromHttpHeader(data *bytes.Buffer) string {
	// Create a reader for the data buffer.
	reader := bufio.NewReader(data)

	// Read and discard the first line (Request-Line)
	// Request-Line:  GET / HTTP/1.1
	_, err := reader.ReadString('\n')
	if err != nil {
		log.Debugf("[HTTP] cannot read request line: %v", err)
		return ""
	}

	// Attempt to parse HTTP headers.
	tp := textproto.NewReader(reader)
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return ""
	}

	// If it is a valid HTTP request, locate the Host header.
	host, ok := headers["Host"]
	if !ok {
		return ""
	}

	// The Host header may include the port number. Split it off.
	domain := strings.Split(host[0], ":")[0]

	return domain
}

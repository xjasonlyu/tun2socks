package socks5

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSocks5ClientHandshake(t *testing.T) {
	// Mock server responses
	readBuffer := &bytes.Buffer{}
	readBuffer.Write([]byte{Version, MethodUserPass})
	readBuffer.Write([]byte{Version, 0x00 /* STATUS of SUCCESS */})
	readBuffer.Write([]byte{Version, 0x00 /* STATUS of SUCCESS */, 0x00 /* RSV */})
	readBuffer.Write([]byte{AtypIPv4, 0x1, 0x2, 0x3, 0x4, 0x0, 0x0 /* IPv4: 1.2.3.4:0 */})
	reader := bufio.NewReader(bytes.NewReader(readBuffer.Bytes()))

	writeBuffer := &bytes.Buffer{}
	writer := bufio.NewWriter(writeBuffer)

	io := bufio.NewReadWriter(reader, writer)

	addr, err := ClientHandshake(io, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, CmdConnect, &User{
		Username: "test",
		Password: "6ab49d8b-a009-44e4-bd53-fbdb48fbe7eb",
	})

	assert.Nil(t, err, "Failed to perform SOCKS5 client handshake: %v", err)
	assert.Equal(t, "1.2.3.4:0", addr.String(), "Incorrect address obtained from SOCKS5 client handshake")
}

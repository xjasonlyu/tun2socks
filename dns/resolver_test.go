package dns

import (
	"net"
	"testing"
	"time"
)

func TestDNSConfiguration(t *testing.T) {
	// Test initial state - DNS should be disabled
	if IsDNSEnabled() {
		t.Error("DNS should be disabled initially")
	}

	// Test enabling DNS with address
	config := &Config{
		Address: "8.8.8.8:53",
	}
	SetConfig(config)

	if !IsDNSEnabled() {
		t.Error("DNS should be enabled when address is set")
	}

	retrievedConfig := GetConfig()
	if retrievedConfig == nil {
		t.Error("Config should not be nil")
		return
	}

	if retrievedConfig.Address != "8.8.8.8:53" {
		t.Errorf("Expected address 8.8.8.8:53, got %s", retrievedConfig.Address)
	}

	// Test disabling DNS
	SetConfig(nil)
	if IsDNSEnabled() {
		t.Error("DNS should be disabled when config is nil")
	}

	// Test with empty address
	emptyConfig := &Config{
		Address: "",
	}
	SetConfig(emptyConfig)
	if IsDNSEnabled() {
		t.Error("DNS should be disabled when address is empty")
	}
}

func TestIsDNSRequest(t *testing.T) {
	// Test DNS port detection
	if !IsDNSRequest(53) {
		t.Error("Port 53 should be detected as DNS")
	}

	if IsDNSRequest(80) {
		t.Error("Port 80 should not be detected as DNS")
	}

	if IsDNSRequest(443) {
		t.Error("Port 443 should not be detected as DNS")
	}
}

// Mock DNS server for testing
func createMockDNSServer(t *testing.T, protocol string) (string, func()) {
	var listener net.Listener
	var packetConn net.PacketConn
	var err error

	if protocol == "tcp" {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to create mock TCP DNS server: %v", err)
		}

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return // Server closed
				}
				go handleMockTCPDNSQuery(conn)
			}
		}()

		return listener.Addr().String(), func() { listener.Close() }
	} else {
		packetConn, err = net.ListenPacket("udp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to create mock UDP DNS server: %v", err)
		}

		go handleMockUDPDNSQueries(packetConn)

		return packetConn.LocalAddr().String(), func() { packetConn.Close() }
	}
}

func handleMockTCPDNSQuery(conn net.Conn) {
	defer conn.Close()

	// Read any data (simplified for testing)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	// Create mock DNS response (just echo back with response flag set)
	if n >= 3 {
		buf[2] |= 0x80 // Set QR bit to indicate response
	}

	// Send response back
	conn.Write(buf[:n])
}

func handleMockUDPDNSQueries(conn net.PacketConn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return // Connection closed
		}

		// Create mock DNS response
		response := createMockDNSResponse(buf[:n])
		conn.WriteTo(response, addr)
	}
}

func createMockDNSResponse(query []byte) []byte {
	// Create a simple mock DNS response
	// This is a minimal response that changes the QR bit to 1 (response)
	if len(query) < 12 {
		return query // Invalid query
	}

	response := make([]byte, len(query))
	copy(response, query)

	// Set QR bit (bit 15 of flags) to 1 to indicate this is a response
	response[2] |= 0x80

	// Set response code to NOERROR (0)
	response[3] = (response[3] & 0xF0) // Clear response code bits

	return response
}

func TestForwardDNSOverTCP(t *testing.T) {
	// Test with DNS hijacking disabled
	SetConfig(nil)

	// Create a mock client connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	err := ForwardDNSOverTCP(clientConn, "example.com:53")
	if err != nil {
		t.Errorf("ForwardDNSOverTCP should return nil when DNS hijacking is disabled, got: %v", err)
	}

	// Test with DNS hijacking enabled
	mockAddr, cleanup := createMockDNSServer(t, "tcp")
	defer cleanup()

	SetConfig(&Config{Address: mockAddr})

	// Create test DNS query (without length prefix for pipe)
	testQuery := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Flags: standard query
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0
		// Query for "test"
		0x04, 't', 'e', 's', 't',
		0x00,       // End of name
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
	}

	// Test the forwarding in a goroutine
	done := make(chan error, 1)
	responseReceived := make(chan bool, 1)

	// Start forwarding
	go func() {
		done <- ForwardDNSOverTCP(serverConn, "test.com:53")
	}()

	// Send query and read response
	go func() {
		// Send query
		clientConn.Write(testQuery)

		// Read response
		response := make([]byte, len(testQuery))
		n, err := clientConn.Read(response)
		if err == nil && n > 0 {
			responseReceived <- true
		} else {
			responseReceived <- false
		}
	}()

	// Wait for either completion or timeout
	select {
	case <-responseReceived:
		// Good, we got a response
	case <-time.After(500 * time.Millisecond):
		t.Error("ForwardDNSOverTCP did not receive response in time")
	}

	// Close connections to end the forwarding
	clientConn.Close()
	serverConn.Close()

	// Wait for forwarding to complete
	select {
	case err := <-done:
		// Error is expected due to connection close
		_ = err
	case <-time.After(500 * time.Millisecond):
		t.Error("ForwardDNSOverTCP did not complete in time")
	}
}

func TestForwardDNSOverUDP(t *testing.T) {
	// Test with DNS hijacking disabled
	SetConfig(nil)

	// Create a mock UDP connection
	mockConn := &mockPacketConn{responses: make(chan []byte, 1)}
	mockAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	testQuery := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Flags: standard query
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0
		// Query for "test"
		0x04, 't', 'e', 's', 't',
		0x00,       // End of name
		0x00, 0x01, // Type: A
		0x00, 0x01, // Class: IN
	}

	err := ForwardDNSOverUDP(mockConn, mockAddr, testQuery)
	if err != nil {
		t.Errorf("ForwardDNSOverUDP should return nil when DNS hijacking is disabled, got: %v", err)
	}

	// Test with DNS hijacking enabled
	mockDNSAddr, cleanup := createMockDNSServer(t, "udp")
	defer cleanup()

	SetConfig(&Config{Address: mockDNSAddr})

	err = ForwardDNSOverUDP(mockConn, mockAddr, testQuery)
	if err != nil {
		t.Errorf("ForwardDNSOverUDP failed: %v", err)
	}

	// Check if response was written back
	select {
	case response := <-mockConn.responses:
		if len(response) == 0 {
			t.Error("Expected DNS response, got empty response")
		}
		// Verify it's a response (QR bit set)
		if len(response) >= 3 && (response[2]&0x80) == 0 {
			t.Error("Expected DNS response flag to be set")
		}
	case <-time.After(1 * time.Second):
		t.Error("ForwardDNSOverUDP did not write response back to client")
	}
}

// Mock PacketConn for testing
type mockPacketConn struct {
	responses chan []byte
}

func (m *mockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return 0, nil, net.ErrClosed // Not used in our test
}

func (m *mockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	response := make([]byte, len(p))
	copy(response, p)
	select {
	case m.responses <- response:
	default:
	}
	return len(p), nil
}

func (m *mockPacketConn) Close() error                       { return nil }
func (m *mockPacketConn) LocalAddr() net.Addr                { return nil }
func (m *mockPacketConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockPacketConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockPacketConn) SetWriteDeadline(t time.Time) error { return nil }

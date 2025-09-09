package dns

import (
	"testing"
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

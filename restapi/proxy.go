package restapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gopkg.in/yaml.v3"
)

var (
	proxyMu     sync.RWMutex
	proxyConfig *ProxyConfig
	// UpdateProxyFunc is a callback function to update proxy in engine
	UpdateProxyFunc func(addr, user, pass string) error
)

func init() {
	// Initialize with safe defaults
	proxyConfig = &ProxyConfig{
		Type:    "socks5",
		Address: "127.0.0.1:7891",
	}
	registerEndpoint("/api/v1/proxy", proxyRouter())
}

func proxyRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getProxyConfig)
	r.Post("/", updateProxyConfig)
	return r
}

type ProxyConfig struct {
	Type     string `json:"type" yaml:"type"`                             // socks5, socks4, http, https
	Address  string `json:"address" yaml:"proxy"`                         // 127.0.0.1:7891
	Username string `json:"username,omitempty" yaml:"username,omitempty"` // optional
	Password string `json:"password,omitempty" yaml:"password,omitempty"` // optional
}

func getProxyConfig(w http.ResponseWriter, r *http.Request) {
	proxyMu.RLock()
	defer proxyMu.RUnlock()

	// Always return the current config, with safe defaults
	response := struct {
		Type    string `json:"type"`
		Address string `json:"address"`
	}{
		Type:    proxyConfig.Type,
		Address: proxyConfig.Address,
	}

	render.JSON(w, r, struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{
		Success: true,
		Message: "Proxy config retrieved",
		Data:    response,
	})
}

func updateProxyConfig(w http.ResponseWriter, r *http.Request) {
	var req ProxyConfig
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrBadRequest)
		return
	}

	proxyMu.Lock()
	proxyConfig = &req
	proxyMu.Unlock()

	// Update proxy dynamically in engine if callback is registered
	if UpdateProxyFunc != nil {
		if err := UpdateProxyFunc(req.Type+"://"+req.Address, req.Username, req.Password); err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, newError("Failed to update proxy: "+err.Error()))
			return
		}
	}

	if err := saveConfigToFile(&req); err != nil {
	}

	render.JSON(w, r, struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Proxy config saved successfully (requires restart)",
	})
}

func saveConfigToFile(config *ProxyConfig) error {
	configPath := getConfigFilePath()

	var existingConfig map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, &existingConfig); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	if existingConfig == nil {
		existingConfig = make(map[string]interface{})
	}
	existingConfig["proxy"] = config.Type + "://" + config.Address
	if config.Username != "" {
		existingConfig["username"] = config.Username
	}
	if config.Password != "" {
		existingConfig["password"] = config.Password
	}

	data, err := yaml.Marshal(existingConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func getConfigFilePath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "tun2socks")
	}
	return filepath.Join(configDir, "config.yaml")
}

func SetProxyConfig(proxyAddr string) {
	proxyMu.Lock()
	defer proxyMu.Unlock()

	if proxyAddr == "" {
		return
	}

	proxyParts := strings.SplitN(proxyAddr, "://", 2)
	if len(proxyParts) == 2 {
		proxyConfig = &ProxyConfig{
			Type:    proxyParts[0],
			Address: proxyParts[1],
		}
	}
}

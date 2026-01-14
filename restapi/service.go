package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

func init() {
	registerEndpoint("/api/v1/service", serviceRouter())
}

func serviceRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getServiceStatus)
	r.Get("/events", serviceEvents)
	r.Post("/", startService)
	r.Delete("/", stopService)
	return r
}

type ServiceStatus struct {
	Running     bool          `json:"running"`
	PID         int           `json:"pid"`
	Uptime      int64         `json:"uptime"`
	Connections int           `json:"connections"`
	MemoryUsage int64         `json:"memoryUsage"`
	CPUUsage    float64       `json:"cpuUsage"`
	Proxy       string        `json:"proxy"`
	Traffic     *TrafficStats `json:"traffic"`
}

type TrafficStats struct {
	UploadBytes   int64 `json:"uploadBytes"`
	DownloadBytes int64 `json:"downloadBytes"`
	UploadSpeed   int64 `json:"uploadSpeed"`
	DownloadSpeed int64 `json:"downloadSpeed"`
}

type startServiceRequest struct {
	Proxy string `json:"proxy"`
}

var (
// serviceStartTime time.Time
)

func getServiceStatus(w http.ResponseWriter, r *http.Request) {
	running := IsEngineRunning()
	proxy := getCurrentProxyConfig()
	startTime := GetEngineStartTime()

	status := ServiceStatus{
		Running: running,
		Proxy:   proxy.Address,
		Traffic: &TrafficStats{},
	}

	if running && !startTime.IsZero() {
		status.Uptime = int64(time.Since(startTime).Seconds())
		status.PID = os.Getpid()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		status.MemoryUsage = int64(m.Alloc)

		upSpeed, downSpeed := statistic.DefaultManager.Now()
		snapshot := statistic.DefaultManager.Snapshot()

		status.Connections = len(snapshot.Connections)

		status.Traffic = &TrafficStats{
			UploadSpeed:   upSpeed,
			DownloadSpeed: downSpeed,
			UploadBytes:   snapshot.UploadTotal,
			DownloadBytes: snapshot.DownloadTotal,
		}
	}

	render.JSON(w, r, render.M{
		"success": true,
		"message": "Service status retrieved",
		"data":    status,
	})
}

func serviceEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			running := IsEngineRunning()
			startTime := GetEngineStartTime()

			status := ServiceStatus{
				Running: running,
				Proxy:   getCurrentProxyConfig().Address,
				Traffic: &TrafficStats{},
			}

			if running && !startTime.IsZero() {
				status.Uptime = int64(time.Since(startTime).Seconds())
				status.PID = os.Getpid()

				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				status.MemoryUsage = int64(m.Alloc)

				upSpeed, downSpeed := statistic.DefaultManager.Now()
				snapshot := statistic.DefaultManager.Snapshot()

				status.Connections = len(snapshot.Connections)

				status.Traffic = &TrafficStats{
					UploadSpeed:   upSpeed,
					DownloadSpeed: downSpeed,
					UploadBytes:   snapshot.UploadTotal,
					DownloadBytes: snapshot.DownloadTotal,
				}
			}

			data, _ := json.Marshal(status)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func startService(w http.ResponseWriter, r *http.Request) {
	var req startServiceRequest
	if r.ContentLength > 0 {
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrBadRequest)
			return
		}
	}

	if IsEngineRunning() {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, newError("Service already running"))
		return
	}

	// serviceStartTime = time.Now()

	render.JSON(w, r, render.M{
		"success": true,
		"message": "Service started successfully",
		"data": ServiceStatus{
			Running: true,
			Uptime:  0,
			PID:     os.Getpid(),
			Traffic: &TrafficStats{},
			Proxy:   getCurrentProxyConfig().Address,
		},
	})
}

func stopService(w http.ResponseWriter, r *http.Request) {
	if !IsEngineRunning() {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, newError("Service not running"))
		return
	}

	// serviceStartTime = time.Time{}

	render.JSON(w, r, render.M{
		"success": true,
		"message": "Service stopped successfully",
		"data":    struct{}{},
	})
}

func getCurrentProxyConfig() *ProxyConfig {
	proxyMu.RLock()
	defer proxyMu.RUnlock()
	return proxyConfig
}

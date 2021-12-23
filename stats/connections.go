package stats

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
)

const defaultInterval = 1000

func init() {
	addMountPoint("/connections", connectionRouter())
}

func connectionRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getConnections)
	r.Delete("/", closeAllConnections)
	r.Delete("/{id}", closeConnection)
	return r
}

func getConnections(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		snapshot := statistic.DefaultManager.Snapshot()
		render.JSON(w, r, snapshot)
		return
	}

	conn, err := _upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	intervalStr := r.URL.Query().Get("interval")
	interval := defaultInterval
	if intervalStr != "" {
		t, err := strconv.Atoi(intervalStr)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrBadRequest)
			return
		}

		interval = t
	}

	buf := &bytes.Buffer{}
	sendSnapshot := func() error {
		buf.Reset()
		snapshot := statistic.DefaultManager.Snapshot()
		if err := json.NewEncoder(buf).Encode(snapshot); err != nil {
			return err
		}

		return conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}

	if err := sendSnapshot(); err != nil {
		return
	}

	tick := time.NewTicker(time.Millisecond * time.Duration(interval))
	defer tick.Stop()
	for range tick.C {
		if err := sendSnapshot(); err != nil {
			break
		}
	}
}

func closeConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	snapshot := statistic.DefaultManager.Snapshot()
	for _, c := range snapshot.Connections {
		if id == c.ID() {
			_ = c.Close()
			break
		}
	}
	render.NoContent(w, r)
}

func closeAllConnections(w http.ResponseWriter, r *http.Request) {
	snapshot := statistic.DefaultManager.Snapshot()
	for _, c := range snapshot.Connections {
		_ = c.Close()
	}
	render.NoContent(w, r)
}

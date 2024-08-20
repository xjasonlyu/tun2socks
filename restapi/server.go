package restapi

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"

	V "github.com/xjasonlyu/tun2socks/v2/internal/version"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

var (
	_upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	_endpoints = make(map[string]http.Handler)
)

func registerEndpoint(pattern string, handler http.Handler) {
	_endpoints[pattern] = handler
}

func Start(addr, token string) error {
	r := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         300,
	})

	r.Use(c.Handler)
	r.Group(func(r chi.Router) {
		r.Use(authenticator(token))
		r.Get("/", hello)
		r.Get("/traffic", traffic)
		r.Get("/version", version)
		// attach HTTP handlers
		for pattern, handler := range _endpoints {
			r.Mount(pattern, handler)
		}
	})

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return http.Serve(listener, r)
}

func hello(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, render.M{"hello": V.Name})
}

func authenticator(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Browser websocket not support custom header
			if websocket.IsWebSocketUpgrade(r) && r.URL.Query().Get("token") != "" {
				t := r.URL.Query().Get("token")
				if t != token {
					render.Status(r, http.StatusUnauthorized)
					render.JSON(w, r, ErrUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")
			text := strings.SplitN(header, " ", 2)

			hasInvalidHeader := text[0] != "Bearer"
			hasInvalidToken := len(text) != 2 || text[1] != token
			if hasInvalidHeader || hasInvalidToken {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func traffic(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		wsConn *websocket.Conn
	)
	if websocket.IsWebSocketUpgrade(r) {
		wsConn, err = _upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
	}

	if wsConn == nil {
		w.Header().Set("Content-Type", "application/json")
		render.Status(r, http.StatusOK)
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	buf := &bytes.Buffer{}
	for range tick.C {
		buf.Reset()

		up, down := statistic.DefaultManager.Now()
		if err = json.NewEncoder(buf).Encode(struct {
			Up   int64 `json:"up"`
			Down int64 `json:"down"`
		}{
			Up:   up,
			Down: down,
		}); err != nil {
			break
		}

		if wsConn == nil {
			_, err = w.Write(buf.Bytes())
			w.(http.Flusher).Flush()
		} else {
			err = wsConn.WriteMessage(websocket.TextMessage, buf.Bytes())
		}

		if err != nil {
			break
		}
	}
}

func version(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, render.M{
		"version": V.Version,
		"commit":  V.GitCommit,
		"modules": V.Info(),
	})
}

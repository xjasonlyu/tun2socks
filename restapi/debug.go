//go:build debug

package restapi

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

func init() {
	registerEndpoint("/debug/pprof/", pprofRouter())
}

func pprofRouter() http.Handler {
	r := chi.NewRouter()
	r.HandleFunc("/", pprof.Index)
	r.HandleFunc("/cmdline", pprof.Cmdline)
	r.HandleFunc("/profile", pprof.Profile)
	r.HandleFunc("/symbol", pprof.Symbol)
	r.HandleFunc("/trace", pprof.Trace)
	r.HandleFunc("/{name}", pprofHandler)
	return r
}

func pprofHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	pprof.Handler(name).ServeHTTP(w, r)
}

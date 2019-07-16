// +build debug

// To view all available profiles, open http://localhost:6060/debug/pprof/ in your browser.

package main

import (
	"net/http"
	_ "net/http/pprof"
	"runtime/debug"
)

func init() {
	// cgo calls will consume more system threads, better keep an eye on that.
	debug.SetMaxThreads(35)

	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()
}

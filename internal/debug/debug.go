// Package debug indicates if the debug tag is enabled at build time.
package debug

var _debug = false

func Debug() bool {
	return _debug
}

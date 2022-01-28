// Package automaxprocs automatically sets GOMAXPROCS to match the Linux
// container CPU quota, if any.
package automaxprocs

import (
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	maxprocs.Set(maxprocs.Logger(func(string, ...interface{}) {}))
}

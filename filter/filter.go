package filter

import (
	"io"
)

// Filter is used for filtering IP packets coming from TUN.
type Filter interface {
	io.Writer
}

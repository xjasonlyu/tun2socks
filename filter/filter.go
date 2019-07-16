package filter

import (
	"io"
)

// Filter is used for filtering IP packets comming from TUN.
type Filter interface {
	io.Writer
}

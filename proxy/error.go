package proxy

import "fmt"

// indicate remote unreachable
type UnreachableError struct {
	proto   string
	code    int
	message string
}

func (e *UnreachableError) Error() string {
	return fmt.Sprintf("UnreachableError proto:%s,code:%d,message:%s", e.proto, e.code, e.message)
}

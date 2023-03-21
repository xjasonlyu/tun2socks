package socks5

import "fmt"

type ReplyRepError struct {
	Command Command
	Rep     ReplyRep
}

func (e *ReplyRepError) Error() string {
	return fmt.Sprintf("ReplyRepError command:%s,rep:%s", e.Command, e.Rep)
}

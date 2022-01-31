package log

import (
	"fmt"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/common/observable"
)

var (
	_logCh  = make(chan interface{})
	_source = observable.NewObservable(_logCh)
)

type Event struct {
	Level   Level     `json:"level"`
	Message string    `json:"msg"`
	Time    time.Time `json:"time"`
}

func newEvent(level Level, format string, args ...interface{}) *Event {
	event := &Event{
		Level:   level,
		Time:    time.Now(),
		Message: fmt.Sprintf(format, args...),
	}
	_logCh <- event /* send all events to logCh */

	return event
}

func Subscribe() observable.Subscription {
	sub, _ := _source.Subscribe()
	return sub
}

func UnSubscribe(sub observable.Subscription) {
	_source.UnSubscribe(sub)
}

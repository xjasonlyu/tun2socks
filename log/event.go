package log

import (
	"fmt"
	"time"

	"github.com/xjasonlyu/tun2socks/common/observable"
)

var (
	logCh  = make(chan interface{})
	source = observable.NewObservable(logCh)
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
	logCh <- event /* send all events to logCh */

	return event
}

func Subscribe() observable.Subscription {
	sub, _ := source.Subscribe()
	return sub
}

func UnSubscribe(sub observable.Subscription) {
	source.UnSubscribe(sub)
}

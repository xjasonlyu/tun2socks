package log

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	level = INFO
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type Event struct {
	LogLevel Level
	Payload  string
}

func (e *Event) Type() string {
	return e.LogLevel.String()
}

func Infof(format string, v ...interface{}) {
	event := newLog(INFO, format, v...)
	printf(event)
}

func Warnf(format string, v ...interface{}) {
	event := newLog(WARNING, format, v...)
	printf(event)
}

func Errorf(format string, v ...interface{}) {
	event := newLog(ERROR, format, v...)
	printf(event)
}

func Debugf(format string, v ...interface{}) {
	event := newLog(DEBUG, format, v...)
	printf(event)
}

func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

func Access(process, outbound, network, local, target string) {
	Infof("[%v] [%v] [%v] %s --> %s", outbound, network, process, local, target)
}

func SetLevel(newLevel Level) {
	level = newLevel
}

func printf(data *Event) {
	if data.LogLevel < level {
		return
	}

	switch data.LogLevel {
	case INFO:
		log.Infoln(data.Payload)
	case WARNING:
		log.Warnln(data.Payload)
	case ERROR:
		log.Errorln(data.Payload)
	case DEBUG:
		log.Debugln(data.Payload)
	}
}

func newLog(logLevel Level, format string, v ...interface{}) *Event {
	return &Event{
		LogLevel: logLevel,
		Payload:  fmt.Sprintf(format, v...),
	}
}

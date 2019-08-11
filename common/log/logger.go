package log

type Level uint8

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	SILENT
)

type Logger interface {
	SetLevel(level Level)
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
}

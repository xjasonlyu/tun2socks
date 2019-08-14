package log

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	SILENT
)

type Level int

func (l Level) String() string {
	switch l {
	case INFO:
		return "info"
	case WARNING:
		return "warning"
	case ERROR:
		return "error"
	case DEBUG:
		return "debug"
	case SILENT:
		return "silent"
	default:
		return "unknown"
	}
}

package log

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Level uint32

const (
	SilentLevel Level = iota
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

// UnmarshalJSON deserialize Level with json
func (level *Level) UnmarshalJSON(data []byte) error {
	var lvl string
	if err := json.Unmarshal(data, &lvl); err != nil {
		return err
	}

	l, err := ParseLevel(lvl)
	if err != nil {
		return err
	}

	*level = l
	return nil
}

// MarshalJSON serialize Level with json
func (level Level) MarshalJSON() ([]byte, error) {
	return json.Marshal(level.String())
}

func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case SilentLevel:
		return "silent"
	default:
		return fmt.Sprintf("not a valid level %d", level)
	}
}

func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "silent":
		return SilentLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	default:
		return Level(0), fmt.Errorf("not a valid logrus Level: %q", lvl)
	}
}

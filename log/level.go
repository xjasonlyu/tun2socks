package log

import (
	"go.uber.org/zap/zapcore"
)

// Level is an alias for zapcore.Level.
type Level = zapcore.Level

// Levels are aliases for Level.
const (
	DebugLevel   = zapcore.DebugLevel
	InfoLevel    = zapcore.InfoLevel
	WarnLevel    = zapcore.WarnLevel
	ErrorLevel   = zapcore.ErrorLevel
	DPanicLevel  = zapcore.DPanicLevel
	PanicLevel   = zapcore.PanicLevel
	FatalLevel   = zapcore.FatalLevel
	InvalidLevel = zapcore.InvalidLevel
	SilentLevel  = InvalidLevel + 1
)

// ParseLevel is a thin wrapper for zapcore.ParseLevel.
func ParseLevel(text string) (Level, error) {
	switch text {
	case "silent", "SILENT":
		return SilentLevel, nil
	default:
		return zapcore.ParseLevel(text)
	}
}

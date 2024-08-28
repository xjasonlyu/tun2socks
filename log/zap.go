package log

import (
	"go.uber.org/zap"
)

// Must is an alias for zap.Must.
var Must = zap.Must

// logger aliases for zap.Logger and zap.SugaredLogger.
type (
	Logger        = zap.Logger
	SugaredLogger = zap.SugaredLogger
)

type (
	// Option is an alias for zap.Option.
	Option = zap.Option
)

// pkgCallerSkip skips the pkg wrapper code as the caller.
var pkgCallerSkip = zap.AddCallerSkip(2)

package pluginutils

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"io"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapWrap simplifies wrapping zap instance to hclog.Logger interfacs.
func ZapWrap(zap *zap.Logger) hclog.Logger {
	return ZapWrapper{Zap: zap}
}

type Level = hclog.Level

// ZapWrapper holds *zap.Logger and adapts its methods to declared by hclog.Logger.
type ZapWrapper struct {
	Zap  *zap.Logger
	name string
}

func (w ZapWrapper) GetLevel() hclog.Level {
	return hclog.NoLevel
}

func (w ZapWrapper) Warn(msg string, args ...interface{}) {
	w.Zap.Warn(msg, convertToZapAny(args...)...)
}
func (w ZapWrapper) Debug(msg string, args ...interface{}) {
	// w.Zap.Debug(msg, convertToZapAny(args...)...)
	w.Zap.Info(msg, convertToZapAny(args...)...)
}
func (w ZapWrapper) Info(msg string, args ...interface{}) {
	w.Zap.Info(msg, convertToZapAny(args...)...)
}
func (w ZapWrapper) Error(msg string, args ...interface{}) {
	w.Zap.Error(msg, convertToZapAny(args...)...)
}

// Log logs messages with four simplified levels - Debug,Warn,Error and Info as a default.
func (w ZapWrapper) Log(lvl Level, msg string, args ...interface{}) {
	switch lvl {
	case hclog.Debug:
		w.Debug(msg, args...)
	case hclog.Warn:
		w.Warn(msg, args...)
	case hclog.Error:
		w.Error(msg, args...)
	case hclog.DefaultLevel, hclog.Info, hclog.NoLevel, hclog.Trace:
		w.Info(msg, args...)
	}
}

// Trace will log an info-level message in Zap.
func (w ZapWrapper) Trace(msg string, args ...interface{}) {
	w.Zap.Info(msg, convertToZapAny(args...)...)
}

// With returns a logger with always-presented key-value pairs.
func (w ZapWrapper) With(args ...interface{}) hclog.Logger {
	return &ZapWrapper{Zap: w.Zap.With(convertToZapAny(args...)...)}
}

// Named returns a logger with the specific nams.
// The name string will always be presented in a log messages.
func (w ZapWrapper) Named(name string) hclog.Logger {
	return &ZapWrapper{Zap: w.Zap.Named(name), name: name}
}

// Name returns a logger's name (if presented).
func (w ZapWrapper) Name() string { return w.name }

// ResetNamed has the same implementation as Named.
func (w ZapWrapper) ResetNamed(name string) hclog.Logger {
	return &ZapWrapper{Zap: w.Zap.Named(name), name: name}
}

// StandardWriter returns os.Stderr as io.Writer.
func (w ZapWrapper) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return hclog.DefaultOutput
}

// StandardLogger returns standard logger with os.Stderr as a writer.
func (w ZapWrapper) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(w.StandardWriter(opts), "", log.LstdFlags)
}

// IsTrace has no implementation.
func (w ZapWrapper) IsTrace() bool { return false }

// IsDebug has no implementation.
func (w ZapWrapper) IsDebug() bool { return false }

// IsInfo has no implementation.
func (w ZapWrapper) IsInfo() bool { return false }

// IsWarn has no implementation.
func (w ZapWrapper) IsWarn() bool { return false }

// IsError has no implementation.
func (w ZapWrapper) IsError() bool { return false }

// ImpliedArgs has no implementation.
func (w ZapWrapper) ImpliedArgs() []interface{} { return nil }

// SetLevel has no implementation.
func (w ZapWrapper) SetLevel(lvl Level) {
}

// Assumed, we'll get key-value pairs as arguments.
// Code below prevents a panic, if wrong arguments set received.
func convertToZapAny(args ...interface{}) []zapcore.Field {
	fields := []zapcore.Field{}
	for i := len(args); i > 0; i -= 2 {
		left := i - 2
		if left < 0 {
			left = 0
		}

		items := args[left:i]

		switch l := len(items); l {
		case 2:
			k, ok := items[0].(string)
			if ok {
				fields = append(fields, zap.Any(k, items[1]))
			} else {
				fields = append(fields, zap.Any(fmt.Sprintf("arg%d", i-1), items[1]))
				fields = append(fields, zap.Any(fmt.Sprintf("arg%d", left), items[0]))
			}
		case 1:
			fields = append(fields, zap.Any(fmt.Sprintf("arg%d", left), items[0]))
		}
	}

	return fields
}

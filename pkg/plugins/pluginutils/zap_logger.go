package pluginutils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/hashicorp/go-hclog"

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
	//return hclog.LevelFromString(w.Zap.Level().String())
}

func (w ZapWrapper) Warn(msg string, args ...interface{}) {
	w.printConvertMessage(hclog.Warn, msg, args...)
}

func (w ZapWrapper) Debug(msg string, args ...interface{}) {
	w.printConvertMessage(hclog.Info, msg, args...) // DEBUG set to Info since we want any logs to appear
}

func (w ZapWrapper) Info(msg string, args ...interface{}) {
	w.printConvertMessage(hclog.Info, msg, args...)
}

func (w ZapWrapper) Error(msg string, args ...interface{}) {
	w.printConvertMessage(hclog.Error, msg, args...)
}

// Log logs messages with four simplified levels - Debug,Warn,Error and Info as a default.
func (w ZapWrapper) Log(lvl Level, msg string, args ...interface{}) {
	w.printConvertMessage(lvl, msg, args...)
}

// Trace will log an info-level message in Zap.
func (w ZapWrapper) Trace(msg string, args ...interface{}) {
	w.printConvertMessage(hclog.Info, msg, args...)
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
	var fields []zapcore.Field
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

// printConvertMessage is a wrapper for hclog messages
func (w ZapWrapper) printConvertMessage(lvl Level, msg string, args ...interface{}) {
	// separator is \t
	split := strings.Split(msg, "\t")
	if len(split) <= 3 {
		w.zapLog(lvl, msg, args...)
		return
	}
	// First is date: ignore
	// Second is level.. ignore?
	// Third is log source
	logSource := split[2]
	// Fourth is the message
	logMsg := split[3]
	// then we come to args, basically it's a key value json string to decode
	var zapArgs []zapcore.Field
	zapArgs = append(zapArgs, zap.String("custom_caller", logSource))
	zapArgs = append(zapArgs, convertToZapAny(args...)...)

	if len(split) > 4 {
		logArgs := map[string]interface{}{}

		// Fifth is the json key value object
		err := json.Unmarshal([]byte(split[4]), &logArgs)
		if err != nil {
			zap.L().Warn("could not unmarshal incoming zap log")
			zapArgs = append(zapArgs, zap.String("custom_args", split[4]))
		} else {
			for key, val := range logArgs {
				zapArgs = append(zapArgs, zap.Any(key, val))
			}
		}
	}

	// Extra args should also be added
	if len(split) > 5 {
		zapArgs = append(zapArgs, zap.Strings("custom_extra_args", split[5:len(split)-1]))
	}

	switch lvl {
	case hclog.Debug:
		w.Zap.Debug(logMsg, zapArgs...)
	case hclog.Warn:
		w.Zap.Warn(logMsg, zapArgs...)
	case hclog.Error:
		w.Zap.Error(logMsg, zapArgs...)
	default:
		w.Zap.Info(logMsg, zapArgs...)
	}
}

// zapLog logs message and converts args to zap args
func (w ZapWrapper) zapLog(lvl Level, msg string, args ...interface{}) {
	switch lvl {
	case hclog.Debug:
		w.Zap.Debug(msg, convertToZapAny(args)...)
	case hclog.Warn:
		w.Zap.Warn(msg, convertToZapAny(args)...)
	case hclog.Error:
		w.Zap.Error(msg, convertToZapAny(args)...)
	default:
		w.Zap.Info(msg, convertToZapAny(args)...)
	}
}

package router

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	gorillacontext "github.com/gorilla/context"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

var (
	// RequestLogger is called by the Logger middleware handler to log each request.
	// Its made a package-level variable so that it can be reconfigured for custom
	// logging configurations.
	RequestLogger = CustomRequestLogger(&CustomLogFormatter{Logger: log.New(os.Stdout, "", log.LstdFlags), NoColor: false})
)

// CustomLogger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return. When standard output is a TTY, Logger will
// print in color, otherwise it will print in black and white. Logger prints a
// request ID if one is provided.
//
// Alternatively, look at https://github.com/pressly/lg and the `lg.RequestLogger`
// middleware pkg.
func CustomLogger(next http.Handler) http.Handler {
	return RequestLogger(next)
}

// CustomRequestLogger returns a logger handler using a custom LogFormatter.
func CustomRequestLogger(f chimiddleware.LogFormatter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := f.NewLogEntry(r)
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				user := gorillacontext.Get(r, models.UserLogin)
				gorillacontext.Clear(r)
				if user != nil {
					cW(entry.(*customLogEntry).buf, !f.(*CustomLogFormatter).NoColor, nGreen, "%s", user)
					entry.(*customLogEntry).buf.WriteString(" - ")
				}

				entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t1), nil)

			}()
			ctx := context.WithValue(r.Context(), models.ContextKeyLoggerR, r)
			next.ServeHTTP(ww, chimiddleware.WithLogEntry(r.WithContext(ctx), entry))
		}
		return http.HandlerFunc(fn)
	}
}

// CustomLogFormatter is a simple logger that implements a LogFormatter.
type CustomLogFormatter struct {
	Logger  chimiddleware.LoggerInterface
	NoColor bool
}

// NewLogEntry creates a new LogEntry for the request.
func (l *CustomLogFormatter) NewLogEntry(r *http.Request) chimiddleware.LogEntry {
	useColor := !l.NoColor
	entry := &customLogEntry{
		CustomLogFormatter: l,
		request:            r,
		buf:                &bytes.Buffer{},
		useColor:           useColor,
	}

	reqID := chimiddleware.GetReqID(r.Context())
	if reqID != "" {
		cW(entry.buf, useColor, nYellow, "[%s] ", reqID)
	}
	cW(entry.buf, useColor, nCyan, "\"")
	cW(entry.buf, useColor, bMagenta, "%s ", r.Method)

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	cW(entry.buf, useColor, nCyan, "%s://%s%s %s\" ", scheme, r.Host, r.RequestURI, r.Proto)

	entry.buf.WriteString("from ")
	entry.buf.WriteString(r.RemoteAddr)
	entry.buf.WriteString(" - ")

	return entry
}

type customLogEntry struct {
	*CustomLogFormatter
	request  *http.Request
	buf      *bytes.Buffer
	useColor bool
}

func (l *customLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	switch {
	case status < 200:
		cW(l.buf, l.useColor, bBlue, "%03d", status)
	case status < 300:
		cW(l.buf, l.useColor, bGreen, "%03d", status)
	case status < 400:
		cW(l.buf, l.useColor, bCyan, "%03d", status)
	case status < 500:
		cW(l.buf, l.useColor, bYellow, "%03d", status)
	default:
		cW(l.buf, l.useColor, bRed, "%03d", status)
	}

	cW(l.buf, l.useColor, bBlue, " %dB", bytes)

	l.buf.WriteString(" in ")
	if elapsed < 500*time.Millisecond {
		cW(l.buf, l.useColor, nGreen, "%s", elapsed)
	} else if elapsed < 5*time.Second {
		cW(l.buf, l.useColor, nYellow, "%s", elapsed)
	} else {
		cW(l.buf, l.useColor, nRed, "%s", elapsed)
	}

	l.Logger.Print(l.buf.String())
}

func (l *customLogEntry) Panic(v interface{}, stack []byte) {
	panicEntry := l.NewLogEntry(l.request).(*customLogEntry)
	cW(panicEntry.buf, l.useColor, bRed, "panic: %+v", v)
	l.Logger.Print(panicEntry.buf.String())
	l.Logger.Print(string(stack))
}

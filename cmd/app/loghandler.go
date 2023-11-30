// Taken from 
// https://github.com/jba/slog/blob/main/handlers/loghandler/log_handler.go
// because Go doesn't give us access to via default handler, thanks Go.
package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const logBufSize = 1024 // the log buffer size, the larger it is, means more memory usage.
// might need larger for the massive discord debug logs.

type logHandler struct {
	opts      LogOptions
	prefix    string // preformatted group names followed by a dot
	preformat string // preformatted Attrs, with an initial space

	mu sync.Mutex
	w  io.Writer
}

//apparently can't embed the slog.HandlerOptions for whatever reason.
type LogOptions struct {
	AddSource bool
	Level slog.Leveler
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	Domain string
}

func NewLogHandler(w io.Writer, opts *LogOptions) *logHandler {
	h := &logHandler{w: w}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.ReplaceAttr == nil {
		h.opts.ReplaceAttr = func(_ []string, a slog.Attr) slog.Attr { return a }
	}
	return h
}

func (h *logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	return &logHandler{
		w:         h.w,
		opts:      h.opts,
		preformat: h.preformat,
		prefix:    h.prefix + name + ".",
	}
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	bufp := allocLogBuf()
	buf := *bufp
	defer func() {
		*bufp = buf
		freeLogBuf(bufp)
	}()

	for _, a := range attrs {
		buf = h.appendAttr(buf, h.prefix, a)
	}

	return &logHandler{
		w:         h.w,
		opts:      h.opts,
		prefix:    h.prefix,
		preformat: h.preformat + string(buf),
	}
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	bufp := allocLogBuf()
	buf := *bufp
	defer func() {
		*bufp = buf
		freeLogBuf(bufp)
	}()

	if !r.Time.IsZero() {
		buf = append(buf, []byte("time=")...)
		buf = r.Time.AppendFormat(buf, time.RFC3339)
		buf = append(buf, ' ')
	}

	buf = append(buf, []byte("level=")...)
	buf = append(buf, r.Level.String()...)
	buf = append(buf, ' ')

	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()

		buf = append(buf, []byte("source=")...)
		buf = append(buf, f.File...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(f.Line), 10)
		buf = append(buf, ' ')
	}

	if h.opts.Domain != "" {
		buf = append(buf, []byte("domain=")...)
		buf = append(buf, h.opts.Domain...)
		buf = append(buf, ' ')
	}

	buf = append(buf, []byte("msg=")...)
	buf = append(buf, r.Message...)

	buf = append(buf, h.preformat...)
	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, h.prefix, a)
		return true
	})

	buf = append(buf, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.w.Write(buf)
	return err
}

func (h *logHandler) appendAttr(buf []byte, prefix string, a slog.Attr) []byte {
	if a.Equal(slog.Attr{}) {
		return buf
	}

	if a.Value.Kind() != slog.KindGroup {
		buf = append(buf, ' ')
		buf = append(buf, prefix...)
		buf = append(buf, a.Key...)
		buf = append(buf, '=')
		return fmt.Appendf(buf, "%v", a.Value.Any())
	}

	// Group
	if a.Key != "" {
		prefix += a.Key + "."
	}
	for _, a := range a.Value.Group() {
		buf = h.appendAttr(buf, prefix, a)
	}

	return buf
}

var logBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, logBufSize)
		return &b
	},
}

func allocLogBuf() *[]byte {
	return logBufPool.Get().(*[]byte)
}

func freeLogBuf(b *[]byte) {
	// To reduce peak allocation, return only smaller buffers to the pool.
	const maxBufferSize = 16 << 10

	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		logBufPool.Put(b)
	}
}
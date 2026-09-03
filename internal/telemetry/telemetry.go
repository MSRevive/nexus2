package telemetry

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/msrevive/nexus2/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/titpetric/oida"
)

var (
	clockMu sync.Mutex
	clockLast time.Time
)

/* ---
	Monotonic clock
	time.Now() on Windows is coarse enough that a fast operation can start and
	finish on the same reading, leaving the span with Duration == 0. oida's
	timeline treats a zero duration as "still open" (frontend/view/timeline.go
	renders the word "open" instead of a time), so completed spans on quick work
	like RunGC, SyncToDisk or a 404 look like they never ended.

	Handing oida a clock that never returns the same time twice keeps every span
	at 1ns or more, which is the difference between "0s" and "open" on the
	dashboard. Drop this once oida checks Ended() there instead.
--- */
func monotonicNow() time.Time {
	clockMu.Lock()
	defer clockMu.Unlock()

	now := time.Now()
	if !now.After(clockLast) {
		now = clockLast.Add(time.Nanosecond)
	}
	clockLast = now

	return now
}

func New(cfg *config.Config, logger *slog.Logger) (*oida.Tracer, error) {
	if !cfg.Telemetry.Enabled {
		return nil, nil
	}

	opts := oida.NewOptions("nexus2")
	opts.Enabled = true

	// Anything left at zero in the config keeps oida's default.
	if cfg.Telemetry.Path != "" {
		opts.Path = cfg.Telemetry.Path
	}
	if cfg.Telemetry.RingBufferSize > 0 {
		opts.RingBufferSize = cfg.Telemetry.RingBufferSize
	}
	if cfg.Telemetry.TopRequests > 0 {
		opts.TopRequests = cfg.Telemetry.TopRequests
	}
	if cfg.Telemetry.MaxSpansPerTrace > 0 {
		opts.MaxSpansPerTrace = cfg.Telemetry.MaxSpansPerTrace
	}
	if cfg.Telemetry.SampleRate > 0 {
		opts.SampleRate = cfg.Telemetry.SampleRate
	}

	// oida already excludes its own dashboard subtree, so only ours need adding.
	opts.IgnorePaths = append(opts.IgnorePaths, cfg.Telemetry.IgnorePaths...)

	// Never report the same instant twice, so a fast span can't measure as zero.
	opts.Clock = monotonicNow

	// oida falls back to r.Pattern, which only net/http's ServeMux sets. Without this
	// every steamid and uuid becomes its own group in the statistics.
	opts.RouteFunc = func(r *http.Request) string {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			return rctx.RoutePattern()
		}
		return ""
	}

	// oida never writes to stdout or stderr, this is the only way to see its failures.
	opts.OnError = func(err error) {
		if logger != nil {
			logger.Error("telemetry error", "error", err)
		}
	}

	return oida.New(opts)
}

func Path(t *oida.Tracer) string {
	if t == nil {
		return ""
	}

	return t.Options().Path
}

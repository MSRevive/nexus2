package telemetry

import (
	"log/slog"
	"net/http"

	"github.com/msrevive/nexus2/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/titpetric/oida"
)
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

	// Don't trace the dashboard itself, it would fill the ring buffer with views of the ring buffer.
	opts.IgnorePaths = append(opts.IgnorePaths, opts.Path, opts.Path+"/*")
	opts.IgnorePaths = append(opts.IgnorePaths, cfg.Telemetry.IgnorePaths...)

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

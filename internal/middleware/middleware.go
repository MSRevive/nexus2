package middleware

import (
	"time"
	"net/http"
	"log/slog"

	"github.com/msrevive/nexus2/internal/config"

	"github.com/saintwish/kv/ccmap"
)

type Middleware struct {
	logger *slog.Logger
	config *config.Config
	ipList *ccmap.Cache[string, string]
}

func New(log *slog.Logger, cfg *config.Config, ipList *ccmap.Cache[string, string]) *Middleware {
	return &Middleware{
		logger: log,
		config: cfg,
		ipList: ipList,
	}
}

func (m *Middleware) Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		// Maximum age allowable under Chromium v76 is 2 hours, so just use that since
		// anything higher will be ignored (even if other browsers do allow higher values).
		//
		// @see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age#Directives
		w.Header().Set("Access-Control-Max-Age", "7200")
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		m.logger.Info("HTTP", "method", r.Method, "URI", r.RequestURI, "IP", r.RemoteAddr, "size", r.ContentLength, "ping", time.Since(start))
	})
}

func (m *Middleware) PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				m.logger.Error("Recovered from fatal error", "error", p)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}
package middleware

import (
	"time"
	"net/http"
	"log/slog"
	"runtime/debug"

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE")
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
		t1 := time.Now()
		defer func() {
			m.logger.Info("", "method", r.Method, "proto", r.Proto, "URI", r.RequestURI, "IP", r.RemoteAddr, "bytes", r.ContentLength, "ping", time.Since(t1))
		}()
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				m.logger.Error("FATALITY", "error", p, "stack", string(debug.Stack()))
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}
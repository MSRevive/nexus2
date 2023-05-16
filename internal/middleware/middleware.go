package middleware

import (
	"time"
	"net/http"

	"github.com/msrevive/nexus2/cmd/app"
	"github.com/msrevive/nexus2/pkg/helper"
)

type Middleware struct {
	app *app.App
}

func New(a *app.App) *Middleware {
	return &Middleware{
		app: a,
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
		m.app.LogAPI.Printf("%s %s from %s (%v) size (%d)", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start), r.ContentLength)
	})
}

func (m *Middleware) PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				m.app.LogAPI.Tracef("Fatal Error: %+v", p)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

/* no authentication 
  Does not do any authentication
---*/
func (mw *Middleware) NoAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
		return
	}
}

/* Level 1 authentication 
  Performs IP whitelist and API key checks against what's allowed (if they're enabled in the config).
  This should be used as the basic authentication
---*/
func (mw *Middleware) Lv1Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := helper.GetIP(r)
		key := r.Header.Get("Authorization")
		
		val,ok := mw.app.IPList[ip]
		if mw.app.Config.ApiAuth.EnforceIP {
			if !ok {
				mw.app.LogAPI.Printf("%s is not authorized.", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		if mw.app.Config.ApiAuth.EnforceKey {
			if !ok || val != key {
				mw.app.LogAPI.Printf("%s failed API key check.", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
		return
	}
}

/* Level 2 authentication 
  Performs level 1 authentication and user agent check.
  This should be used to make sure the request came from msr game server.
---*/
func (mw *Middleware) Lv2Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := helper.GetIP(r)
		key := r.Header.Get("Authorization")
		
		val,ok := mw.app.IPList[ip]
		if mw.app.Config.ApiAuth.EnforceIP {
			if !ok {
				mw.app.LogAPI.Printf("%s is not authorized.", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		if mw.app.Config.ApiAuth.EnforceKey {
			if !ok || val != key {
				mw.app.LogAPI.Printf("%s failed API key check.", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}
		
		//if useragent in config is empty then just skip.
		if mw.app.Config.Verify.Useragent != "" {
			if r.UserAgent() != mw.app.Config.Verify.Useragent {
				mw.app.LogAPI.Printf("%s incorrect user agent!", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
		return
	}
}
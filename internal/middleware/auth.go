package middleware

import (
	"net/http"

	"github.com/msrevive/nexus2/pkg/helper"
)


/* ---
	Tier 0 Authenication
	Performs no authenication.
--- */
func (mw *Middleware) Tier0Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

/* ---
	Tier 1 Authenication
	Performs IP whitelist and API key checks against what's allowed (if they're enabled in the config).
  	This should be used as the basic authentication
--- */
func (mw *Middleware) Tier1Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := helper.GetIP(r)
		key := r.Header.Get("Authorization")
		
		val,ok := mw.a.List.IP.GetHas(ip)
		if mw.a.Config.ApiAuth.EnforceIP {
			if !ok {
				mw.a.Logger.API.Info("IP is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		if mw.a.Config.ApiAuth.EnforceKey {
			if !ok || val != key {
				mw.a.Logger.API.Info("API key is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

/* ---
	Tier 2 Authenication
	Performs level 1 authentication and user agent check.
  	This should be used to make sure the request came from msr game server.
--- */
func (mw *Middleware) Tier2Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := helper.GetIP(r)
		key := r.Header.Get("Authorization")
		
		val,ok := mw.a.List.IP.GetHas(ip)
		if mw.a.Config.ApiAuth.EnforceIP {
			if !ok {
				mw.a.Logger.API.Info("IP is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		if mw.a.Config.ApiAuth.EnforceKey {
			if !ok || val != key {
				mw.a.Logger.API.Info("API key is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}
		
		//if useragent in config is empty then just skip.
		if mw.a.Config.Verify.Useragent != "" {
			if r.UserAgent() != mw.a.Config.Verify.Useragent {
				mw.a.Logger.API.Info("Incorrect user agent!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
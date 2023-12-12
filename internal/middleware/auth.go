package middleware

import (
	"net/http"

	"github.com/msrevive/nexus2/pkg/utils"
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
		ip := utils.GetIP(r)
		key := r.Header.Get("Authorization")
		
		val,ok := mw.ipList.GetHas(ip)
		if mw.config.ApiAuth.EnforceIP {
			if !ok {
				mw.logger.Info("IP is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		if mw.config.ApiAuth.EnforceKey {
			if !ok || val != key {
				mw.logger.Info("API key is not authorized!", "ip", ip)
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
		ip := utils.GetIP(r)
		key := r.Header.Get("Authorization")
		
		val,ok := mw.ipList.GetHas(ip)
		if mw.config.ApiAuth.EnforceIP {
			if !ok {
				mw.logger.Info("IP is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		if mw.config.ApiAuth.EnforceKey {
			if !ok || val != key {
				mw.logger.Info("API key is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}
		
		//if useragent in config is empty then just skip.
		if mw.config.Verify.Useragent != "" {
			if r.UserAgent() != mw.config.Verify.Useragent {
				mw.logger.Info("Incorrect user agent!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
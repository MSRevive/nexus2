package middleware

import (
	"net/http"

	"github.com/msrevive/nexus2/pkg/utils"
)


/* ---
	Basic authenication
	Performs no authenication.
--- */
func (mw *Middleware) BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

/* ---
	External Authenication
	Performs IP whitelist and API key checks against what's allowed (if they're enabled in the config).

	SystemAdmins [IP address] = API key check for external endpoints.
--- */
func (mw *Middleware) ExternalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetIP(r)
		key := r.Header.Get("Authorization")

		val,ok := mw.systemAdmins.GetHas(ip)

		if !ok || (key != "" && val != key) {
			mw.logger.Info("ExternalAuth: IP is not authorized!", "ip", ip)
			http.Error(w, http.StatusText(401), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

/* ---
	Internal Authenication
	Performs IP whitelist and API key checks against what's allowed (if they're enabled in the config).
  	These requests should only come from MSR game server.
--- */
func (mw *Middleware) InternalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetIP(r)
		key := r.Header.Get("Authorization")
		
		val,ok := mw.ipList.GetHas(ip)
		if mw.config.ApiAuth.EnforceIP {
			if !ok {
				mw.logger.Info("InternalAuth: IP is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		if mw.config.ApiAuth.EnforceKey {
			if !ok || val != key {
				mw.logger.Info("InternalAuth: API key is not authorized!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}
		
		//if useragent in config is empty then just skip.
		if mw.config.ApiAuth.UserAgent != "" {
			if r.UserAgent() != mw.config.ApiAuth.UserAgent {
				mw.logger.Info("InternalAuth: Incorrect user agent!", "ip", ip)
				http.Error(w, http.StatusText(401), http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
package middleware

import (
	"net/http"

	"github.com/msrevive/nexus2/cmd/app"
)

func setControlHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	// Maximum age allowable under Chromium v76 is 2 hours, so just use that since
	// anything higher will be ignored (even if other browsers do allow higher values).
	//
	// @see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age#Directives
	w.Header().Set("Access-Control-Max-Age", "7200")
}

//false = not allowed, true = is allowed
func checkIP(ip string, a *app.App) bool {
	if a.Config.ApiAuth.EnforceIP {      
		if _,ok := a.IPList[ip]; !ok {
			return false
		}
		
		return true
	}
	
	return true
}

//false = not allowed, true = is allowed
func checkAPIKey(key string, a *app.App) bool {
	if a.Config.ApiAuth.EnforceKey {
		if key == a.Config.ApiAuth.Key {
			return true
		}
		
		return false
	}
	
	return true
}
package middleware

import(
	"net"
	"strings"
	"net/http"
	
	"github.com/msrevive/nexus2/internal/system"
)

func getIP(r *http.Request) string {
	ip := r.Header.Get("X_Real_IP")
	if ip == "" {
		ips := strings.Split(r.Header.Get("X_Forwarded_For"), ", ")
		if ips[0] != "" {
			 return ips[0]
		}

		ip,_,_ = net.SplitHostPort(r.RemoteAddr)
		return ip
	}

	return ip
}

func setControlHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	// Maximum age allowable under Chromium v76 is 2 hours, so just use that since
	// anything higher will be ignored (even if other browsers do allow higher values).
	//
	// @see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age#Directives
	w.Header().Set("Access-Control-Max-Age", "7200")
}

//false = not allowed, true = is allowed
func checkIP(ip string) bool {
	if system.AuthCfg.IsEnforcingIP() {      
		if !system.AuthCfg.IsKnownIP(ip) {
			return false
		}
		
		return true
	}
	
	return true
}

//false = not allowed, true = is allowed
func checkAPIKey(key string) bool {
	if system.AuthCfg.IsEnforcingKey() {
		if system.AuthCfg.IsValidKey(key) {
			return true
		}
		
		return false
	}
	
	return true
}
package middleware

import (
	"github.com/msrevive/nexus2/cmd/app"
)

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
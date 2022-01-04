package middleware

import(
  "net"
  "net/http"
  
  "github.com/msrevive/nexus2/session"
  "github.com/msrevive/nexus2/log"
  "github.com/msrevive/nexus2/rate"
)

var (
  globalLimiter *rate.Limiter
)

func getIP(req *http.Request) string {
  ip := req.Header.Get("X_Real_IP")
  if ip == "" {
    ips := strings.Split(req.Header.Get("X_Forwarded_For"), ", ")
    if ips[0] != "" {
       return ips[0]
    }

    ip,_,_ = net.SplitHostPort(req.RemoteAddr)
    return ip
  }

  return ip
}

func Init(next http.HandlerFunc) http.HandlerFunc {
  return func(res http.ResponseWriter, req *http.Request){
    ip := getIP(req)

    log.Log.Errorf("Request sent from %s", ip)
    
    //IP Auth
    if session.Config.APIAuth.EnforceIP {
      if _,ok := session.Config.APIAuth.IPAllowed[ip]; !ok {
        log.Log.Errorf("%s Is not authorized.", ip)
        http.Error(res, http.StatusText(401), http.StatusUnauthorized)
        return
      }
    }
    
    //API Key Auth
    if session.Config.APIAuth.EnforceKey {
      if req.Header.Get("Authorization") != session.Config.ApiAuth.Key {
        log.Log.Errorf("%s failed API key check.", ip)
        http.Error(res, http.StatusText(401), http.StatusUnauthorized)
        return
      }
    }

    if globalLimiter == nil {
      log.Log.Debugln("Creating global rate limiter.")
      globalLimiter = rate.NewLimiter(1, session.Config.Core.MaxRequests, session.Config.Core.MaxAge, 0)
    }

    globalLimiter.CheckTime()
    if globalLimiter.IsAllowed() == false {
      log.Log.Println("Too many global requests sent.")
      http.Error(res, http.StatusText(429), http.StatusTooManyRequests)
      return
    }

    next(res, req)
    return
  }
}
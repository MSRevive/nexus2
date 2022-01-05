package middleware

import(
  "net"
  "time"
  "strings"
  "net/http"
  "runtime/debug"
  
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

func Log(next http.Handler) http.Handler {
  return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
    start := time.Now()
    log.Log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
    next.ServeHTTP(res, req)
  })
}

func PanicRecovery(next http.Handler) http.Handler {
  return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
    defer func() {
      if panic := recover(); panic != nil {
        http.Error(res, http.StatusText(501), http.StatusInternalServerError)
        log.Log.Errorln("501: We have encountered an error with the last request.")
        log.Log.Errorf("501: Error: %s", panic.(error).Error())
        log.Log.Errorf(string(debug.Stack()))
      }
    }()
    next.ServeHTTP(res, req)
  })
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
  return func(res http.ResponseWriter, req *http.Request) {
    ip := getIP(req)
    
    //IP Auth
    if session.Config.ApiAuth.EnforceIP {
      if _,ok := session.IPList[ip]; !ok {
        log.Log.Printf("%s Is not authorized.", ip)
        http.Error(res, http.StatusText(401), http.StatusUnauthorized)
        return
      }
    }
    
    //API Key Auth
    if session.Config.ApiAuth.EnforceKey {
      if req.Header.Get("Authorization") != session.Config.ApiAuth.Key {
        log.Log.Printf("%s failed API key check.", ip)
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
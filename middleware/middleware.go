package middleware

import(
  "net"
  "time"
  "strings"
  "net/http"
  "runtime/debug"
  
  "github.com/msrevive/nexus2/system"
  "github.com/msrevive/nexus2/log"
  "github.com/msrevive/nexus2/rate"
)

var (
  globalLimiter *rate.Limiter
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

func Log(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    setControlHeaders(w) //best place to set control headers?
    start := time.Now()
    next.ServeHTTP(w, r)
    log.Log.Printf("%s %s from %s (%v)", r.Method, r.RequestURI, getIP(r), time.Since(start))
  })
}

func PanicRecovery(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if panic := recover(); panic != nil {
        http.Error(w, http.StatusText(501), http.StatusInternalServerError)
        log.Log.Errorln("501: We have encountered an error with the last request.")
        log.Log.Errorf("501: Error: %s", panic.(error).Error())
        log.Log.Errorf(string(debug.Stack()))
      }
    }()
    
    next.ServeHTTP(w, r)
  })
}

func RateLimit(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if globalLimiter == nil {
      globalLimiter = rate.NewLimiter(1, system.Config.RateLimit.MaxRequests, system.Config.RateLimit.MaxAge, 0)
    }

    globalLimiter.CheckTime()
    if globalLimiter.IsAllowed() == false {
      log.Log.Println("Received too many requests.")
      http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
      return
    }
    
    next.ServeHTTP(w, r)
  })
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    ip := getIP(r)
    
    //IP Auth
    if system.Config.ApiAuth.EnforceIP {      
      if _,ok := system.IPList[ip]; !ok {
        log.Log.Printf("%s Is not authorized.", ip)
        http.Error(w, http.StatusText(401), http.StatusUnauthorized)
        return
      }
    }
    
    //API Key Auth
    if system.Config.ApiAuth.EnforceKey {
      if r.Header.Get("Authorization") != system.Config.ApiAuth.Key {
        log.Log.Printf("%s failed API key check.", ip)
        http.Error(w, http.StatusText(401), http.StatusUnauthorized)
        return
      }
    }

    next(w, r)
    return
  }
}

func NoAuth(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    next(w, r)
    return
  }
}
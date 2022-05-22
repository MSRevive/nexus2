package middleware

import(
  "time"
  "net/http"
  "runtime/debug"
  
  "github.com/msrevive/nexus2/system"
  "github.com/msrevive/nexus2/log"
  "github.com/msrevive/nexus2/rate"
)

var (
  globalLimiter *rate.Limiter
)

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
        http.Error(w, http.StatusText(500), http.StatusInternalServerError)
        log.Log.Errorf("HTTP Error (500): %s", panic.(error).Error())
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

/* no authentication 
  Does not do any authentication
---*/
func NoAuth(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    next(w, r)
    return
  }
}

/* Level 1 authentication 
  Performs IP whitelist and API key checks against what's allowed (if they're enabled in the config).
  This should be used as the basic authentication
---*/
func Lv1Auth(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    ip := getIP(r)
    key := r.Header.Get("Authorization")
    
    //IP Auth
    if !checkIP(ip) {
      log.Log.Printf("%s is not authorized.", ip)
      http.Error(w, http.StatusText(401), http.StatusUnauthorized)
      return
    }
    
    //API Key Auth
    if !checkAPIKey(key) {
      log.Log.Printf("%s failed API key check.", ip)
      http.Error(w, http.StatusText(401), http.StatusUnauthorized)
      return
    }

    next(w, r)
    return
  }
}

/* Level 2 authentication 
  Performs level 1 authentication and user agent check.
  This should be used to make sure the request came from msr game server.
---*/
func Lv2Auth(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    ip := getIP(r)
    key := r.Header.Get("Authorization")
    
    //IP Auth
    if !checkIP(ip) {
      log.Log.Printf("%s is not authorized.", ip)
      http.Error(w, http.StatusText(401), http.StatusUnauthorized)
      return
    }
    
    //API Key Auth
    if !checkAPIKey(key) {
      log.Log.Printf("%s failed API key check.", ip)
      http.Error(w, http.StatusText(401), http.StatusUnauthorized)
      return
    }
    
    log.Log.Debugf("Useragent: %s", r.UserAgent())

    next(w, r)
    return
  }
}
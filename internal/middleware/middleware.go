package middleware

import(
  "time"
  "net/http"
  "runtime/debug"
  
  "github.com/msrevive/nexus2/internal/system"
  "github.com/msrevive/nexus2/pkg/rate"

  "github.com/saintwish/auralog"
)

type middleware struct {
	log *auralog.Logger
	globalLimiter *rate.Limiter
}

func New(logger *auralog.Logger) *middleware {
	return &middleware{
		log: logger,
		globalLimiter: rate.NewLimiter(1, system.HelperCfg.GetMaxRequests(), system.HelperCfg.GetMaxAge(), 0),
	}
}

func (mw *middleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setControlHeaders(w) //best place to set control headers?
		start := time.Now()
		next.ServeHTTP(w, r)
		mw.log.Printf("%s %s from %s (%v)", r.Method, r.RequestURI, getIP(r), time.Since(start))
	})
}

func (mw *middleware) PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
		  if panic := recover(); panic != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			mw.log.Errorf("Fatal Error: %s", panic.(error).Error())
			mw.log.Errorf(string(debug.Stack()))
		  }
		}()
		
		next.ServeHTTP(w, r)
	})
}

func (mw *middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.globalLimiter == nil {
		  mw.globalLimiter = rate.NewLimiter(1, system.HelperCfg.GetMaxRequests(), system.HelperCfg.GetMaxAge(), 0)
		}
	
		mw.globalLimiter.CheckTime()
		if mw.globalLimiter.IsAllowed() == false {
		  mw.log.Println("Received too many requests.")
		  http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		  return
		}
		
		next.ServeHTTP(w, r)
	})
}

/* No authentication
  Does not do any authentication
---*/
func (mw *middleware) NoAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
		return
	}
}

/* Level 1 authentication 
  Performs IP whitelist and API key checks against what's allowed (if they're enabled in the config).
  This should be used as the basic authentication
---*/
func (mw *middleware) Lv1Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		key := r.Header.Get("Authorization")
		
		//IP Auth
		if !checkIP(ip) {
		  mw.log.Printf("%s is not authorized.", ip)
		  http.Error(w, http.StatusText(401), http.StatusUnauthorized)
		  return
		}
		
		//API Key Auth
		if !checkAPIKey(key) {
		  mw.log.Printf("%s failed API key check.", ip)
		  http.Error(w, http.StatusText(401), http.StatusUnauthorized)
		  return
		}
	
		next(w, r)
		return
	}
}

/* Level 2 authentication 
  Performs level 1 authentication and user agent check.
  This should be used to make sure the request came from a MSR game server.
---*/
func (mw *middleware) Lv2Auth(next http.HandlerFunc) http.HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		key := r.Header.Get("Authorization")
		
		//IP Auth
		if !checkIP(ip) {
		  mw.log.Printf("%s is not authorized!", ip)
		  http.Error(w, http.StatusText(401), http.StatusUnauthorized)
		  return
		}
		
		//API Key Auth
		if !checkAPIKey(key) {
		  mw.log.Printf("%s failed API key check!", ip)
		  http.Error(w, http.StatusText(401), http.StatusUnauthorized)
		  return
		}
		
		//if useragent in config is empty then just skip.
		if system.AuthCfg.GetUserAgent() != "" {
		  if r.UserAgent() != system.AuthCfg.GetUserAgent() {
			mw.log.Printf("%s incorrect user agent!", ip)
			http.Error(w, http.StatusText(401), http.StatusUnauthorized)
			return
		  }
		}
	
		next(w, r)
		return
	}
}
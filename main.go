package main

import(
  "time"
  "runtime"
  "strconv"
  "context"
  "net/http"
  "crypto/tls"

  "github.com/msrevive/nexus2/session"
  "github.com/msrevive/nexus2/log"
  
  "github.com/gorilla/mux"
)

func main() {
  if err := session.LoadConfig("./runtime/config.toml"); err != nil {
    panic(err)
  }
  
  //Initiate logging
  log.InitLogging("server.log", session.Config.Log.Dir)
  
  //Max threads allowed.
  if session.Config.Core.MaxThreads != 0 {
    runtime.GOMAXPROCS(session.Config.Core.MaxThreads)
  }
  
  //Web server
  var srv *http.Server
  address := session.Config.Core.IP+":"+strconv.Itoa(session.Config.Core.Port)
  router := mux.NewRouter()
  srv = &http.Server{
    Handler: router,
    Addr: address,
    WriteTimeout: 15 * time.Second,
    ReadTimeout: 15 * time.Second,
  }
  ctx, cancel := context.WithTimeout(context.Background(), session.Config.Core.Graceful * time.Second)
  defer cancel()
  
  //doRoutes(router)
  
  if session.Config.TLS.Enable {
    cert,_ := tls.LoadX509KeyPair(session.Config.TLS.CertFile, session.Config.TLS.KeyFile)
    srv.TLSConfig = &tls.Config{
      Certificates: []tls.Certificate{cert},
    }
    log.Log.Printf("Listening on: %v TLS", session.Config.Core.Port)
    if err := srv.ListenAndServeTLS("", ""); err != nil {
      panic(err)
    }
  }else{
    log.Log.Printf("Listening on: %v", session.Config.Core.Port)
    if err := srv.ListenAndServe(); err != nil {
      panic(err)
    }
  }
  
  defer srv.Shutdown(ctx)
}
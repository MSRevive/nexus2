package main

import(
  "os"
  "time"
  "syscall"
  "runtime"
  "strconv"
  "context"
  "net/http"
  "os/signal"

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
  address := session.Config.Core.IP+":"+strconv.Itoa(session.Config.Core.Port)
  router := mux.NewRouter().PathPrefix(session.Config.Core.RootPath)
  srv := &http.Server{
    Handler: router,
    Addr: address,
    WriteTimeout: 15 * time.Second,
    ReadTimeout: 15 * time.Second,
  }
  ctx, cancel := context.WithTimeout(context.Background(), session.Config.ApiAuth.Graceful * time.Second)
  defer cancel()
  
  //doRoutes(router)
  
  log.Log.Println("Webserver is now running.")
  if session.Config.TLS.Enable {
    if err := srv.ListenAndServeTLS(session.Config.TLS.Certfile, session.Config.TLS.KeyFile, nil); err != nil {
      panic(err)
    }
  }else{
    if err := srv.ListenAndServe(); err != nil {
      panic(err)
    }
  }
  log.Log.Printf("Listening on %s", session.Config.Core.Port)
  
  defer srv.Shutdown(ctx)
}
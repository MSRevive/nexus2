package main

import(
  "time"
  "runtime"
  "strconv"
  "context"
  "net/http"
  "crypto/tls"

  "github.com/msrevive/nexus2/session"
  "github.com/msrevive/nexus2/middleware"
  "github.com/msrevive/nexus2/controller"
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
  
  //Load json files.
  if session.Config.ApiAuth.
  session.LoadIPList(session.Config.ApiAuth.IPListFile)
  session.LoadMapList(session.Config.Verify.MapListFile)
  session.LoadBanList(session.Config.Verify.BanListFile)
  
  //variables for web server
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
  
  //middleware
  router.Use(middleware.Log)
  router.Use(middleware.PanicRecovery)
  
  //api routes
  apic := controller.New(router.PathPrefix(session.Config.Core.RootPath).Subrouter())
  apic.R.HandleFunc("/", middleware.Auth(apic.TestRoot)).Methods(http.MethodGet)
  apic.R.HandleFunc("/map/{name}/{hash}", middleware.Auth(apic.GetMapVerify)).Methods(http.MethodGet)
  apic.R.HandleFunc("/ban/{steamid}", middleware.Auth(apic.GetBanVerify)).Methods(http.MethodGet)
  apic.R.HandleFunc("/sc/{hash}", middleware.Auth(apic.GetSCVerify)).Methods(http.MethodGet)
  
  //character routes
  charc := controller.New(router.PathPrefix(session.Config.Core.RootPath+"/character").Subrouter())
  charc.R.HandleFunc("/", middleware.Auth(charc.GetAllCharacters)).Methods(http.MethodGet)
  charc.R.HandleFunc("/{steamid}", middleware.Auth(charc.GetCharacters)).Methods(http.MethodGet)
  charc.R.HandleFunc("/{steamid}/{slot}", middleware.Auth(charc.GetCharacter)).Methods(http.MethodGet)
  charc.R.HandleFunc("/id/{uid}", middleware.Auth(charc.GetCharacterByID)).Methods(http.MethodGet)
  charc.R.HandleFunc("/", middleware.Auth(charc.PostCharacter)).Methods(http.MethodPost)
  charc.R.HandleFunc("/{uid}", middleware.Auth(charc.PutCharacter)).Methods(http.MethodPut)
  charc.R.HandleFunc("/{uid}", middleware.Auth(charc.DeleteCharacter)).Methods(http.MethodDelete)
  
  //start the web server
  if session.Config.TLS.Enable {
    cert,_ := tls.LoadX509KeyPair(session.Config.TLS.CertFile, session.Config.TLS.KeyFile)
    srv.TLSConfig = &tls.Config{
      Certificates: []tls.Certificate{cert},
    }
    log.Log.Printf("Listening on: %s TLS", address)
    if err := srv.ListenAndServeTLS("", ""); err != nil {
      panic(err)
    }
  }else{
    log.Log.Printf("Listening on: %s", address)
    if err := srv.ListenAndServe(); err != nil {
      panic(err)
    }
  }
  
  defer srv.Shutdown(ctx)
}
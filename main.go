package main

import(
  "time"
  "runtime"
  "strconv"
  "context"
  "flag"
  "net/http"
  "crypto/tls"

  "github.com/msrevive/nexus2/session"
  "github.com/msrevive/nexus2/middleware"
  "github.com/msrevive/nexus2/controller"
  "github.com/msrevive/nexus2/log"
  "github.com/msrevive/nexus2/ent"
  //_ "github.com/msrevive/nexus2/sqlite3"
  
  "github.com/gorilla/mux"
  "golang.org/x/crypto/acme/autocert"
  _ "github.com/mattn/go-sqlite3"
)

func main() {
  var cdir string
  flag.StringVar(&cdir, "cfile", "./runtime/config.toml", "Where to load the config file.")
  flag.BoolVar(&session.Dbg, "dbg", false, "Run with debug mode.")
  flag.Parse()
  
  if err := session.LoadConfig(cdir); err != nil {
    panic(err)
  }
  
  //Initiate logging
  log.InitLogging("server.log", session.Config.Log.Dir, session.Config.Log.Level)
  
  if session.Dbg {
    log.Log.Warnln("Running in Debug mode, do not use in production!")
  }
  
  //Max threads allowed.
  if session.Config.Core.MaxThreads != 0 {
    runtime.GOMAXPROCS(session.Config.Core.MaxThreads)
  }
  
  //Load json files.
  if session.Config.ApiAuth.EnforceIP {
    log.Log.Printf("Loading IP list from %s", session.Config.ApiAuth.IPListFile)
    if err := session.LoadIPList(session.Config.ApiAuth.IPListFile); err != nil {
      log.Log.Warnln("Failed to load IP list.")
    }
  }
  
  if session.Config.Verify.EnforceMap {
    log.Log.Printf("Loading Map list from %s", session.Config.Verify.MapListFile)
    if err := session.LoadMapList(session.Config.Verify.MapListFile); err != nil {
      log.Log.Warnln("Failed to load Map list.")
    }
  }
  
  if session.Config.Verify.EnforceBan {
    log.Log.Printf("Loading Ban list from %s", session.Config.Verify.BanListFile)
    if err := session.LoadBanList(session.Config.Verify.BanListFile); err != nil {
      log.Log.Warnln("Failed to load Ban list.")
    }
  }
  
  //Connect database.
  log.Log.Println("Connecting to database")
  client, err := ent.Open("sqlite3", session.Config.Core.DBString)
  if err != nil {
    log.Log.Fatalf("failed to open connection to sqlite3: %v", err)
  }
  if err := client.Schema.Create(context.Background()); err != nil {
		log.Log.Fatalf("failed to create schema resources: %v", err)
	}
  session.Client = client
  defer session.Client.Close()
  
  //variables for web server
  var srv *http.Server
  address := session.Config.Core.Address+":"+strconv.Itoa(session.Config.Core.Port)
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
  apic.R.HandleFunc("/ban/{steamid:[0-9]+}", middleware.Auth(apic.GetBanVerify)).Methods(http.MethodGet)
  apic.R.HandleFunc("/sc/{hash}", middleware.Auth(apic.GetSCVerify)).Methods(http.MethodGet)
  
  //character routes
  charc := controller.New(router.PathPrefix(session.Config.Core.RootPath+"/character").Subrouter())
  charc.R.HandleFunc("/", middleware.Auth(charc.GetAllCharacters)).Methods(http.MethodGet)
  charc.R.HandleFunc("/id/{uid}", middleware.Auth(charc.GetCharacterByID)).Methods(http.MethodGet)
  charc.R.HandleFunc("/{steamid:[0-9]+}", middleware.Auth(charc.GetCharacters)).Methods(http.MethodGet)
  charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}", middleware.Auth(charc.GetCharacter)).Methods(http.MethodGet)
  charc.R.HandleFunc("/", middleware.Auth(charc.PostCharacter)).Methods(http.MethodPost)
  charc.R.HandleFunc("/{uid}", middleware.Auth(charc.PutCharacter)).Methods(http.MethodPut)
  charc.R.HandleFunc("/{uid}", middleware.Auth(charc.DeleteCharacter)).Methods(http.MethodDelete)
  
  //start the web server
  if session.Config.Cert.Enable {
    certManager := autocert.Manager{
      Prompt: autocert.AcceptTOS,
      HostPolicy: autocert.HostWhitelist(session.Config.Cert.Domain),
      Cache: autocert.DirCache("./runtime/certs"),
    }
    
    srv.TLSConfig = &tls.Config{
      GetCertificate: certManager.GetCertificate,
    }
    
    go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
    
    log.Log.Printf("Listening on: %s TLS", address)
    if err := srv.ListenAndServeTLS("", ""); err != nil {
      log.Log.Fatalf("failed to serve over HTTPS: %v", err)
    }
  }else{
    log.Log.Printf("Listening on: %s", address)
    if err := srv.ListenAndServe(); err != nil {
      log.Log.Fatalf("failed to serve over HTTP: %v", err)
    }
  }
  
  defer srv.Shutdown(ctx)
}
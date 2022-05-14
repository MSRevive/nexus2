package main

import(
  "time"
  "runtime"
  "strconv"
  "context"
  "flag"
  "fmt"
  "net/http"
  "crypto/tls"

  "github.com/msrevive/nexus2/system"
  "github.com/msrevive/nexus2/middleware"
  "github.com/msrevive/nexus2/controller"
  "github.com/msrevive/nexus2/log"
  "github.com/msrevive/nexus2/ent"
  
  "github.com/gorilla/mux"
  "golang.org/x/crypto/acme"
  "golang.org/x/crypto/acme/autocert"
  "entgo.io/ent/dialect/sql/schema"
  _ "github.com/mattn/go-sqlite3"
)

func initPrint() {
  fmt.Printf(`
    _   __                    ___ 
   / | / /__  _  ____  Nexus2|__ \
  /  |/ / _ \| |/_/ / / / ___/_/ /
 / /|  /  __/>  </ /_/ (__  ) __/ 
/_/ |_/\___/_/|_|\__,_/____/____/ 

Copyright © %d, Team MSRebirth

Version: %s
Website: https://msrebirth.net/
License: GPL-3.0 https://github.com/MSRevive/nexus2/blob/main/LICENSE %s`, time.Now().Year(), system.Version, "\n\n")
}

func main() {
  var cfile string
  flag.StringVar(&cfile, "cfile", "./runtime/config.toml", "Where to load the config file.")
  flag.BoolVar(&system.Dbg, "dbg", false, "Run with debug mode.")
  flag.Parse()
  
  if err := system.LoadConfig(cfile); err != nil {
    panic(err)
  }
  
  //initial print
  initPrint()
  
  //Initiate logging
  log.InitLogging("server.log", system.Config.Log.Dir, system.Config.Log.Level, system.Config.Log.ExpireTime)
  
  if system.Dbg {
    log.Log.Warnln("Running in Debug mode, do not use in production!")
  }
  
  //Max threads allowed.
  if system.Config.Core.MaxThreads != 0 {
    runtime.GOMAXPROCS(system.Config.Core.MaxThreads)
  }
  
  //Load json files.
  if system.Config.ApiAuth.EnforceIP {
    log.Log.Printf("Loading IP list from %s", system.Config.ApiAuth.IPListFile)
    if err := system.LoadIPList(system.Config.ApiAuth.IPListFile); err != nil {
      log.Log.Warnln("Failed to load IP list.")
    }
  }
  
  if system.Config.Verify.EnforceMap {
    log.Log.Printf("Loading Map list from %s", system.Config.Verify.MapListFile)
    if err := system.LoadMapList(system.Config.Verify.MapListFile); err != nil {
      log.Log.Warnln("Failed to load Map list.")
    }
  }
  
  if system.Config.Verify.EnforceBan {
    log.Log.Printf("Loading Ban list from %s", system.Config.Verify.BanListFile)
    if err := system.LoadBanList(system.Config.Verify.BanListFile); err != nil {
      log.Log.Warnln("Failed to load Ban list.")
    }
  }
  
  log.Log.Printf("Loading Admin list from %s", system.Config.Verify.AdminListFile)
  if err := system.LoadAdminList(system.Config.Verify.AdminListFile); err != nil {
    log.Log.Warnln("Failed to load Admin list.")
  }
  
  //Connect database.
  log.Log.Println("Connecting to database")
  client, err := ent.Open("sqlite3", system.Config.Core.DBString)
  if err != nil {
    log.Log.Fatalf("failed to open connection to sqlite3: %v", err)
  }
  if err := client.Schema.Create(context.Background(), schema.WithAtlas(true)); err != nil {
		log.Log.Fatalf("failed to create schema resources: %v", err)
	}
  system.Client = client
  defer system.Client.Close()
  
  //variables for web server
  var srv *http.Server
  router := mux.NewRouter()
  srv = &http.Server{
    Handler: router,
    Addr: system.Config.Core.Address+":"+strconv.Itoa(system.Config.Core.Port),
    WriteTimeout: 15 * time.Second,
    ReadTimeout: 15 * time.Second,
    // DefaultTLSConfig sets sane defaults to use when configuring the internal
    // webserver to listen for public connections.
    //
    // @see https://blog.cloudflare.com/exposing-go-on-the-internet
    // credit to https://github.com/pterodactyl/wings/blob/develop/config/config.go
    TLSConfig: &tls.Config{
      NextProtos: []string{"h2", "http/1.1"},
      CipherSuites: []uint16{
    		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
    		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
    		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
    		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
    		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
    	},
      PreferServerCipherSuites: true,
      MinVersion: tls.VersionTLS12,
      MaxVersion: tls.VersionTLS13,
      CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
    },
  }
  
  //middleware
  router.Use(middleware.PanicRecovery)
  router.Use(middleware.Log)
  if system.Config.RateLimit.Enable {
    router.Use(middleware.RateLimit)
  }
  
  //api routes
  apic := controller.New(router.PathPrefix(system.Config.Core.RootPath).Subrouter())
  apic.R.HandleFunc("/", middleware.Auth(apic.TestRoot)).Methods(http.MethodGet)
  apic.R.HandleFunc("/ping", middleware.Auth(apic.GetPing)).Methods(http.MethodGet)
  apic.R.HandleFunc("/map/{name}/{hash}", middleware.Auth(apic.GetMapVerify)).Methods(http.MethodGet)
  apic.R.HandleFunc("/ban/{steamid:[0-9]+}", middleware.Auth(apic.GetBanVerify)).Methods(http.MethodGet)
  apic.R.HandleFunc("/sc/{hash}", middleware.Auth(apic.GetSCVerify)).Methods(http.MethodGet)
  
  //character routes
  charc := controller.New(router.PathPrefix(system.Config.Core.RootPath+"/character").Subrouter())
  charc.R.HandleFunc("/", middleware.Auth(charc.GetAllCharacters)).Methods(http.MethodGet)
  charc.R.HandleFunc("/id/{uid}", middleware.Auth(charc.GetCharacterByID)).Methods(http.MethodGet)
  charc.R.HandleFunc("/{steamid:[0-9]+}", middleware.Auth(charc.GetCharacters)).Methods(http.MethodGet)
  charc.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}", middleware.Auth(charc.GetCharacter)).Methods(http.MethodGet)
  charc.R.HandleFunc("/export/{steamid:[0-9]+}/{slot:[0-9]}", middleware.Auth(charc.ExportCharacter)).Methods(http.MethodGet)
  charc.R.HandleFunc("/", middleware.Auth(charc.PostCharacter)).Methods(http.MethodPost)
  charc.R.HandleFunc("/{uid}", middleware.Auth(charc.PutCharacter)).Methods(http.MethodPut)
  charc.R.HandleFunc("/{uid}", middleware.Auth(charc.DeleteCharacter)).Methods(http.MethodDelete)
  
  if system.Config.Cert.Enable {
    cm := autocert.Manager{
      Prompt: autocert.AcceptTOS,
      HostPolicy: autocert.HostWhitelist(system.Config.Cert.Domain),
      Cache: autocert.DirCache("./runtime/certs"),
    }
  
    srv.TLSConfig = &tls.Config{
      GetCertificate: cm.GetCertificate,
      NextProtos: append(srv.TLSConfig.NextProtos, acme.ALPNProto), // enable tls-alpn ACME challenges
    }
  
    go func() {
      if err := http.ListenAndServe(":http", cm.HTTPHandler(nil)); err != nil {
        log.Log.Fatalf("failed to serve autocert server: %v", err)
      }
    }()
  
    log.Log.Printf("Listening on: %s TLS", srv.Addr)
    if err := srv.ListenAndServeTLS("", ""); err != nil {
      log.Log.Fatalf("failed to serve over HTTPS: %v", err)
    }
  }else{
    log.Log.Printf("Listening on: %s", srv.Addr)
    if err := srv.ListenAndServe(); err != nil {
      log.Log.Fatalf("failed to serve over HTTP: %v", err)
    }
  }
}
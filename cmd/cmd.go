package cmd

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
	"errors"
	"os/signal"

	"github.com/msrevive/nexus2/cmd/app"
	"github.com/msrevive/nexus2/internal/controller"
	"github.com/msrevive/nexus2/internal/middleware"

	"github.com/saintwish/auralog"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	_ "github.com/mattn/go-sqlite3"
)

var (
	logCore *auralog.Logger // Logs for core/server
	logAPI *auralog.Logger // Logs for endpoints/middleware
)

type flags struct {
	address string
	port int
	configFile string
	migrateConfig bool
	debug bool
}

func doFlags(args []string) *flags {
	flgs := &flags{}

	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flagSet.StringVar(&flgs.address, "addr", "127.0.0.1", "The address of the server.")
	flagSet.IntVar(&flgs.port, "port", 1337, "The port this should run on.")
	flagSet.StringVar(&flgs.configFile, "cfile", "./runtime/config.yaml", "Location of via config file")
	flagSet.BoolVar(&flgs.debug, "d", false, "Run with debug mode.")
	flagSet.BoolVar(&flgs.migrateConfig, "m", false, "Migrate the ini/toml config to YAML")
	flagSet.Parse(args[1:])

	return flgs
}

func initLoggers(filename string, dir string, level string, expire string) {
	ex, _ := time.ParseDuration(expire)
	flags := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsWarn := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsError := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile
	flagsDebug := auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile

	file := &auralog.RotateWriter{
		Dir: dir,
		Filename: filename,
		ExTime: ex,
		MaxSize: 5 * auralog.Megabyte,
	}

	logCore = auralog.New(auralog.Config{
		Output: io.MultiWriter(os.Stdout, file),
		Prefix: "[CORE] ",
		Level: auralog.ToLogLevel(level),
		Flag: flags,
		WarnFlag: flagsWarn,
		ErrorFlag: flagsError,
		DebugFlag: flagsDebug,
	})

	logAPI = auralog.New(auralog.Config{
		Output: io.MultiWriter(os.Stdout, file),
		Prefix: "[API] ",
		Level: auralog.ToLogLevel(level),
		Flag: flags,
		WarnFlag: flagsWarn,
		ErrorFlag: flagsError,
		DebugFlag: flagsDebug,
	})
}

func Run(args []string) error {
	flgs := doFlags(args)

	config, err := app.LoadConfig(flgs.configFile)
	if err != nil {
		return err
	}

	apps := app.New(config);

	if flgs.migrateConfig {
		fmt.Println("Running migration...")
		if err := apps.MigrateConfig(); err != nil {
			fmt.Printf("Migration error: %s", err)
		}
		fmt.Println("Finished migration, starting server...")
	}

	//Initiate logging
	initLoggers("server.log", apps.Config.Log.Dir, apps.Config.Log.Level, apps.Config.Log.ExpireTime)
	apps.SetupLoggers(logCore, logAPI)

	//Max threads allowed.
	if apps.Config.Core.MaxThreads != 0 {
		runtime.GOMAXPROCS(apps.Config.Core.MaxThreads)
	}

	//Load json files.
	if apps.Config.ApiAuth.EnforceIP {
		logCore.Printf("Loading IP list from %s", apps.Config.ApiAuth.IPListFile)
		if err := apps.LoadIPList(apps.Config.ApiAuth.IPListFile); err != nil {
			logCore.Warnln("Failed to load IP list.")
		}
	}

	if apps.Config.Verify.EnforceMap {
		logCore.Printf("Loading Map list from %s", apps.Config.Verify.MapListFile)
		if err := apps.LoadMapList(apps.Config.Verify.MapListFile); err != nil {
			logCore.Warnln("Failed to load Map list.")
		}
	}

	if apps.Config.Verify.EnforceBan {
		logCore.Printf("Loading Ban list from %s", apps.Config.Verify.BanListFile)
		if err := apps.LoadBanList(apps.Config.Verify.BanListFile); err != nil {
			logCore.Warnln("Failed to load Ban list.")
		}
	}

	logCore.Printf("Loading Admin list from %s", apps.Config.Verify.AdminListFile)
	if err := apps.LoadAdminList(apps.Config.Verify.AdminListFile); err != nil {
		logCore.Warnln("Failed to load Admin list.")
	}

	//Connect database.
	logCore.Println("Connecting to database")
	if err := apps.SetupClient(); err != nil {
		return err
	}
	defer apps.Client.Close()

	//variables for web server
	var srv *http.Server
	router := chi.NewRouter()
	srv = &http.Server{
		Handler:      router,
		Addr:         apps.Config.Core.Address + ":" + strconv.Itoa(apps.Config.Core.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
			MinVersion:               tls.VersionTLS12,
			MaxVersion:               tls.VersionTLS13,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		},
	}

	//middleware
	mw := middleware.New(apps)
	router.Use(mw.PanicRecovery)
	router.Use(mw.Log)
	if apps.Config.RateLimit.Enable {
		router.Use(mw.RateLimit)
	}

	con := controller.New(apps)
	router.Route(apps.Config.Core.RootPath, func(r chi.Router) {
		r.Get("/ping/", mw.Lv2Auth(con.GetPing))
		r.Get("/map/{name}/{hash}", mw.Lv1Auth(con.GetMapVerify))
		r.Get("/ban/{steamid:[0-9]+}", mw.Lv1Auth(con.GetBanVerify))
		r.Get("/sc/{hash}", mw.Lv1Auth(con.GetSCVerify))
	})

	router.Route(apps.Config.Core.RootPath+"/character", func(r chi.Router) {
		r.Get("/", mw.Lv1Auth(con.GetAllCharacters))
		r.Get("/id/{uid}", mw.Lv1Auth(con.GetCharacterByID))
		r.Get("/{steamid:[0-9]+}", mw.Lv1Auth(con.GetCharacters))
		r.Get("/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(con.GetCharacter))
		r.Get("/export/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(con.ExportCharacter))

		r.Post("/", mw.Lv2Auth(con.PostCharacter))
		r.Put("/{uid}", mw.Lv2Auth(con.PutCharacter))
		r.Delete("/{uid}", mw.Lv2Auth(con.DeleteCharacter))
		r.Patch("/{uid}/restore", mw.Lv1Auth(con.RestoreCharacter))
		r.Get("/{steamid:[0-9]+}/{slot:[0-9]}/versions", mw.Lv1Auth(con.CharacterVersions))
	})

	router.Route(apps.Config.Core.RootPath+"/character/rollback", func(r chi.Router) {
		r.Patch("/{steamid:[0-9]+}/{slot:[0-9]}/{version:[0-9]+}", mw.Lv1Auth(con.RollbackCharacter))
		r.Patch("/{steamid:[0-9]+}/{slot:[0-9]}/latest", mw.Lv1Auth(con.RollbackLatestCharacter))
		r.Delete("/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(con.DeleteRollbacksCharacter))
	})

	if apps.Config.Cert.Enable {
		cm := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(apps.Config.Cert.Domain),
			Cache:      autocert.DirCache("./runtime/certs"),
		}

		srv.TLSConfig = &tls.Config{
			GetCertificate: cm.GetCertificate,
			NextProtos:     append(srv.TLSConfig.NextProtos, acme.ALPNProto), // enable tls-alpn ACME challenges
		}

		go func() {
			if err := http.ListenAndServe(":http", cm.HTTPHandler(nil)); err != nil {
				logCore.Errorf("failed to serve autocert server: %v", err)
			}
		}()
		
		go func() {
			logCore.Printf("Listening on: %s TLS", srv.Addr)
			if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				errMsg := errors.New(fmt.Sprintf("failed to serve over HTTPS: %v", err))
				panic(errMsg)
			}
		}()
	} else {
		go func() {
			logCore.Printf("Listening on: %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errMsg := errors.New(fmt.Sprintf("failed to serve over HTTP: %v", err))
				panic(errMsg)
			}
		}()
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s

	//wait 5 seconds before timing out
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err // failure/timeout shutting down the server gracefully
	}

	return nil
}

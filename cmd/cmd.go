package cmd

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"time"
	"errors"
	"os/signal"

	"github.com/msrevive/nexus2/cmd/app"
	"github.com/msrevive/nexus2/internal/controller"
	"github.com/msrevive/nexus2/internal/middleware"
	"github.com/msrevive/nexus2/internal/database/mongodb"
	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/saintwish/auralog"
	"github.com/saintwish/auralog/rw"
	"github.com/go-chi/chi/v5"
	cmw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	flag "github.com/spf13/pflag"
)

var (
	logCore *auralog.Logger // Logs for core/server
	logAPI *auralog.Logger // Logs for endpoints/middleware
)

type flags struct {
	cfgFile string
	debug bool
	threads int
}

func doFlags(args []string) *flags {
	flgs := &flags{}

	flagSet := flag.NewFlagSet(args[0], flag.ExitOnError)
	flagSet.StringVarP(&flgs.cfgFile, "config", "c", "./runtime/config.yaml", "Location of via config file")
	flagSet.BoolVarP(&flgs.debug, "debug", "d", false, "Run with debug mode.")
	flagSet.IntVarP(&flgs.threads, "t", "threads", 0, "The maximum number of threads the app is allowed to use.")
	flagSet.Parse(args[1:])

	return flgs
}

func initLoggers(filename string, dir string, level string, expire string) {
	ex, _ := time.ParseDuration(expire)
	flags := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsWarn := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds
	flagsError := auralog.Ldate | auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile
	flagsDebug := auralog.Ltime | auralog.Lmicroseconds | auralog.Lshortfile

	file := &rw.RotateWriter{
		Dir: dir,
		Filename: filename,
		ExpireTime: ex,
		MaxSize: 5 * rwriter.Megabyte,
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

func Run(args []string) (error) {
	flgs := doFlags(args)

	if flgs.debug {
		fmt.Println("!!! Running in Debug mode, do not use in production! !!!")
	}

	//Max threads allowed.
	if flgs.threads != 0 {
		runtime.GOMAXPROCS(flgs.threads)
	}

	/////////////////////////
	//Config and database dependencies.
	/////////////////////////
	config, err := config.LoadConfig(flgs.cfgFile)
	if err != nil {
		return err
	}

	db := mongodb.New()
	a := app.New(config, db);
	a.Debug = flgs.debug

	/////////////////////////
	//Logger Dependency
	/////////////////////////
	initLoggers("server.log", config.Log.Dir, config.Log.Level, config.Log.ExpireTime)
	a.SetupLoggers(logCore, logAPI)

	/////////////////////////
	//Load JSON files into lists
	/////////////////////////
	if config.ApiAuth.EnforceIP {
		fmt.Printf("Loading IP list from %s\n", config.ApiAuth.IPListFile)
		if err := a.LoadIPList(config.ApiAuth.IPListFile); err != nil {
			logCore.Warnln("Failed to load IP list.")
		}
	}

	if config.Verify.EnforceMap {
		fmt.Printf("Loading Map list from %s\n", config.Verify.MapListFile)
		if err := a.LoadMapList(config.Verify.MapListFile); err != nil {
			logCore.Warnln("Failed to load Map list.")
		}
	}

	if config.Verify.EnforceBan {
		fmt.Printf("Loading Ban list from %s\n", config.Verify.BanListFile)
		if err := a.LoadBanList(config.Verify.BanListFile); err != nil {
			logCore.Warnln("Failed to load Ban list.")
		}
	}

	fmt.Printf("Loading Admin list from %s", config.Verify.AdminListFile)
	if err := a.LoadAdminList(config.Verify.AdminListFile); err != nil {
		logCore.Warnln("Failed to load Admin list.")
	}

	/////////////////////////
	//Setup HTTP Server
	/////////////////////////
	router := chi.NewRouter()
	a.SetHTTPServer(&http.Server{
		Handler:      router,
		Addr:         config.Core.Address + ":" + strconv.Itoa(config.Core.Port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
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
	})

	/////////////////////////
	//Middleware
	/////////////////////////
	mw := middleware.New(apps)
	router.Use(cmw.RealIP)
	router.Use(mw.Headers)
	if config.RateLimit.MaxRequests > 0 {
		if dur,err := time.ParseDuration(config.RateLimit.MaxAge); err != nil {
			router.Use(httprate.Limit(
				config.RateLimit.MaxRequests,
				dur,
				httprate.WithKeyFuncs(httprate.KeyByIP),
				httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
					response.TooManyRequests(w)
				})),
			)
		}
	}
	router.Use(mw.Log)
	router.Use(mw.PanicRecovery)
	
	con := controller.New(apps)
	router.Route(app.APIPrefix, func(r chi.Router) {
		r.Get("/ping", mw.Lv2Auth(con.GetPing))
		r.Get("/map/{name}/{hash}", mw.Lv1Auth(con.GetMapVerify))
		r.Get("/ban/{steamid:[0-9]+}", mw.Lv1Auth(con.GetBanVerify))
		r.Get("/sc/{hash}", mw.Lv1Auth(con.GetSCVerify))
		if config.Core.Debug {
			r.Mount("/debug", cmw.Profiler())
		}
	})

	router.Route(app.APIPrefix+"/character", func(r chi.Router) {
		//r.Get("/", mw.Lv1Auth(con.GetAllCharacters))
		r.Get("/id/{uid}", mw.Lv1Auth(con.GetCharacterByID))
		r.Get("/{steamid:[0-9]+}", mw.Lv1Auth(con.GetCharacters))
		r.Get("/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(con.GetCharacter))
		r.Get("/export/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(con.ExportCharacter))

		r.Post("/", mw.Lv2Auth(con.PostCharacter))
		r.Put("/{uid}", mw.Lv2Auth(con.PutCharacter))
		r.Delete("/{uid}", mw.Lv2Auth(con.DeleteCharacter))

		r.Patch("/transfer/{uid}/to/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(con.CharacterTransfer))
		r.Patch("/copy/{uid}/to/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(con.CharacterCopy))
	})

	router.Route(app.APIPrefix+"/character/rollback", func(r chi.Router) {
		r.Patch("/{uid}/restore", mw.Lv1Auth(con.RestoreCharacter))
		r.Patch("/{steamid:[0-9]+}/{slot:[0-9]}/restore", mw.Lv1Auth(con.RestoreCharacterBySteamID))

		r.Get("/{steamid:[0-9]+}/{slot:[0-9]}/versions", mw.Lv1Auth(con.CharacterVersions))
		
		r.Patch("/{steamid:[0-9]+}/{slot:[0-9]}/{version:[0-9]+}", mw.Lv1Auth(con.RollbackCharacter))
		r.Patch("/{steamid:[0-9]+}/{slot:[0-9]}/latest", mw.Lv1Auth(con.RollbackLatestCharacter))
		r.Delete("/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(con.DeleteRollbacksCharacter))
	})

	/////////////////////////
	//Auto certificate
	/////////////////////////
	if err := a.Start(); err != nil {
		a.Logger.Core.Error(err)
		return err
	}
	defer func() {
		if err := a.Close(); err != nil {
			a.Logger.Core.Error(err)
			return err
		}
	}

	fmt.Println("\nNexus2 is now running. Press CTRL-C to exit.\n")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s

	return nil
}

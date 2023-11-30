package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
	"syscall"
	"os/signal"
	"crypto/tls"
	"net/http"

	"github.com/msrevive/nexus2/cmd/app"
	"github.com/msrevive/nexus2/internal/controller"
	"github.com/msrevive/nexus2/internal/middleware"
	"github.com/msrevive/nexus2/internal/database/mongodb"
	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/static"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
	cmw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	flag "github.com/spf13/pflag"
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
	flagSet.IntVarP(&flgs.threads, "threads", "t", 0, "The maximum number of threads the app is allowed to use.")
	flagSet.Parse(args[1:])

	return flgs
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
	//Config
	/////////////////////////
	config, err := config.LoadConfig(flgs.cfgFile)
	if err != nil {
		return err
	}

	/////////////////////////
	//Database
	/////////////////////////
	db := mongodb.New()

	/////////////////////////
	//Application
	/////////////////////////
	a := app.New(config, db);

	/////////////////////////
	//Logger Dependency
	/////////////////////////
	a.InitializeLoggers()

	/////////////////////////
	//Load JSON files into lists
	/////////////////////////
	if config.ApiAuth.EnforceIP {
		fmt.Printf("Loading IP list from %s\n", config.ApiAuth.IPListFile)
		if err := a.LoadIPList(config.ApiAuth.IPListFile); err != nil {
			a.Logger.Core.Warn("Failed to load IP list.")
		}
	}

	if config.Verify.EnforceMap {
		fmt.Printf("Loading Map list from %s\n", config.Verify.MapListFile)
		if err := a.LoadMapList(config.Verify.MapListFile); err != nil {
			a.Logger.Core.Warn("Failed to load Map list.")
		}
	}

	if config.Verify.EnforceBan {
		fmt.Printf("Loading Ban list from %s\n", config.Verify.BanListFile)
		if err := a.LoadBanList(config.Verify.BanListFile); err != nil {
			a.Logger.Core.Warn("Failed to load Ban list.")
		}
	}

	fmt.Printf("Loading Admin list from %s\n", config.Verify.AdminListFile)
	if err := a.LoadAdminList(config.Verify.AdminListFile); err != nil {
		a.Logger.Core.Warn("Failed to load Admin list.")
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
	mw := middleware.New(a)

	/////////////////////////
	//Routing
	/////////////////////////
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
	router.Use(cmw.Timeout(a.Config.Core.Timeout * time.Second))

	con := controller.New(a)
	router.Route(static.APIVersion, func(r chi.Router) {
		r.Route("/internal", func(r chi.Router) {
			r.Use(mw.Tier2Auth)

			r.Get("/map/{name}/{hash}", con.GetMapVerify)
			r.Get("/ban/{steamid:[0-9]+}", con.GetBanVerify)
			r.Get("/sc/{hash}", con.GetSCVerify)
		})

		r.Get("/ping", con.GetPing)
	})

	/////////////////////////
	//Auto certificate
	/////////////////////////
	if err := a.Start(); err != nil {
		a.Logger.Core.Error("Fatal error", "error", err)
		return err
	}

	fmt.Println("\nNexus2 is now running. Press CTRL-C to exit.\n")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s

	if err := a.Close(); err != nil {
		a.Logger.Core.Error("Fatal error", "error", err)
		return err
	}

	return nil
}

package cmd

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"syscall"
	"os/signal"
	"net/http"

	"github.com/msrevive/nexus2/cmd/app"
	"github.com/msrevive/nexus2/internal/controller"
	"github.com/msrevive/nexus2/internal/middleware"
	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/service"
	"github.com/msrevive/nexus2/internal/static"
	"github.com/msrevive/nexus2/internal/database/mongodb"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
	cmw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/spf13/pflag"
)

type flags struct {
	cfgFile string
	debug bool
	threads int
}

func doFlags(args []string) *flags {
	flgs := &flags{}

	flagSet := pflag.NewFlagSet(args[0], pflag.ExitOnError)
	flagSet.StringVarP(&flgs.cfgFile, "config", "c", "./runtime/config.yaml", "Location of via config file")
	flagSet.BoolVarP(&flgs.debug, "debug", "d", false, "Run with debug mode.")
	flagSet.IntVarP(&flgs.threads, "threads", "t", 0, "The maximum number of threads the app is allowed to use.")
	flagSet.Parse(args[1:])

	return flgs
}

func Run(args []string) (error) {
	flags := doFlags(args)

	if flags.debug {
		fmt.Println("!!! Running in Debug mode, do not use in production! !!!")
	}

	//Max threads allowed.
	if flags.threads != 0 {
		runtime.GOMAXPROCS(flags.threads)
	}

	/////////////////////////
	//Config
	/////////////////////////
	config, err := config.LoadConfig(flags.cfgFile)
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
			a.Logger.Warn("Failed to load IP list.", "error", err)
		}
	}

	if config.Verify.EnforceMap {
		fmt.Printf("Loading Map list from %s\n", config.Verify.MapListFile)
		if err := a.LoadMapList(config.Verify.MapListFile); err != nil {
			a.Logger.Warn("Failed to load Map list.", "error", err)
		}
	}

	if config.Verify.EnforceBan {
		fmt.Printf("Loading Ban list from %s\n", config.Verify.BanListFile)
		if err := a.LoadBanList(config.Verify.BanListFile); err != nil {
			a.Logger.Warn("Failed to load Ban list.", "error", err)
		}
	}

	fmt.Printf("Loading Admin list from %s\n", config.Verify.AdminListFile)
	if err := a.LoadAdminList(config.Verify.AdminListFile); err != nil {
		a.Logger.Warn("Failed to load Admin list.", "error", err)
	}

	/////////////////////////
	//Middleware
	/////////////////////////
	mw := middleware.New(a.Logger, a.Config, a.List.IP)

	/////////////////////////
	//Routing
	/////////////////////////
	router := chi.NewRouter()
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

	service := service.New(a.DB, a.Config)
	con := controller.New(a.Logger, a.Config, service, a.List.Ban, a.List.Map, a.List.Admin)
	router.Route(static.APIVersion, func(r chi.Router) {
		// Internal use for the game server only.
		r.Route("/internal", func(r chi.Router) {
			r.Use(mw.Tier2Auth)

			r.Get("/map/{name}/{hash}", con.GetMapVerify)
			r.Get("/ban/{steamid:[0-9]+}", con.GetBanVerify)
			r.Get("/sc/{hash}", con.GetSCVerify)

			r.Route("/character", func(r chi.Router) {
				r.Post("/", con.PostCharacter)
				r.Put("/{uuid}", con.PutCharacter)
				r.Delete("/{uuid}", con.SoftDeleteCharacter)

				r.Get("/{uuid}", con.GetCharacterByID)
				r.Get("/{steamid:[0-9]+}", con.GetCharacters)
				r.Get("/{steamid:[0-9]+}/{slot:[0-9]}", con.GetCharacter)
			})
		})

		r.Route("/", func(r chi.Router) {
			r.Use(mw.Tier1Auth)

			r.Route("/character", func(r chi.Router) {
				r.Get("/lookup/{steamid:[0-9]+}/{slot:[0-9]}", con.LookUpCharacterID)
				r.Get("/deleted/{steamid:[0-9]+}", con.GetDeletedCharacters)
				r.Get("/{steamid:[0-9]+}", con.GetCharacters)
				r.Get("/{uuid}", con.GetCharacterByID)
				r.Patch("/restore/{uuid}", con.RestoreCharacter)
				r.Get("/export/{uuid}", con.ExportCharacter)
			})

			r.Route("/rollback/character", func(r chi.Router) {
				r.Get("/{uuid}", con.GetCharacterVersions)
				r.Patch("/{uuid}/latest", con.RollbackCharToLatest)
				r.Patch("/{uuid}/{version:[0-9]+}", con.RollbackCharToVersion)
				r.Delete("/{uuid}", con.DeleteCharRollbacks)
			})
		})

		r.Route("/unsafe", func(r chi.Router) {
			r.Use(mw.Tier2Auth)

			r.Route("/character", func(r chi.Router) {
				r.Patch("/move/{uuid}/to/{steamid:[0-9]+}/{slot:[0-9]}", con.UnsafeMoveCharacter)
				r.Patch("/copy/{uuid}/to/{steamid:[0-9]+}/{slot:[0-9]}", con.UnsafeCopyCharacter)
				r.Delete("/delete/{uuid}", con.UnsafeDeleteCharacter)
			})
		})

		r.Get("/ping", con.GetPing)
		if flags.debug {
			r.Mount("/debug", cmw.Profiler())
		}
	})

	/////////////////////////
	//Auto certificate
	/////////////////////////
	if err := a.Start(router); err != nil {
		a.Logger.Error("Failed to start application", "error", err)
		return err
	}

	fmt.Println("\nNexus2 is now running. Press CTRL-C to exit.\n")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s

	if err := a.Close(); err != nil {
		a.Logger.Error("Failed to close application", "error", err)
		return err
	}

	return nil
}

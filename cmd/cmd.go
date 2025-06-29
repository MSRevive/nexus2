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
	"github.com/msrevive/nexus2/internal/service"
	"github.com/msrevive/nexus2/internal/static"
	"github.com/msrevive/nexus2/internal/response"

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
	//Application Stuff
	/////////////////////////
	a := app.New();

	if err := a.LoadConfig(flags.cfgFile); err != nil {
		return err
	}

	if err := a.InitializeLogger(); err != nil {
		return err
	}

	if err := a.SetupDatabase(); err != nil {
		return err
	}

	if err := a.LoadLists(); err != nil {
		a.Logger.Warn("Failed to load list(s)!", "error", err)
	}

	/////////////////////////
	//Middleware
	/////////////////////////
	mw := middleware.New(a.Logger, a.Config, a.List.IP, a.List.SystemAdmin)

	/////////////////////////
	//Routing
	/////////////////////////
	router := chi.NewRouter()
	router.Use(cmw.RealIP)
	router.Use(mw.Headers)
	if a.Config.RateLimit.MaxRequests > 0 {
		if dur,err := time.ParseDuration(a.Config.RateLimit.MaxAge); err != nil {
			router.Use(httprate.Limit(
				a.Config.RateLimit.MaxRequests,
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
	con := controller.New(service, a.Logger, a.Config, controller.Options{
		MapList: a.List.Map,
	})

	// API version 2
	router.Route(static.APIVersion, func(r chi.Router) {
		r.Use(mw.BasicAuth)
		
		// Internal use for the game server only.
		r.Route("/internal", func(r chi.Router) {
			if !flags.debug {
				r.Use(mw.InternalAuth)
			}

			r.Get("/map/{name}/{hash}", con.GetMapVerify)
			r.Get("/ban/{steamid:[0-9]+}", con.GetBanVerify)
			r.Get("/sc/{hash}", con.GetSCVerify)
			r.Get("/server/{hash}", con.GetServerVerify)
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				response.OK(w, true)
			})

			r.Route("/character", func(r chi.Router) {
				r.Post("/", con.PostCharacter)
				r.Put("/{uuid}", con.PutCharacter)
				r.Delete("/{uuid}", con.SoftDeleteCharacter)

				r.Get("/{steamid:[0-9]+}/{slot:[0-9]+}", con.GetCharacter)
			})
		})

		r.Route("/", func(r chi.Router) {
			if !flags.debug {
				r.Use(mw.ExternalAuth)
			}

			r.Route("/character", func(r chi.Router) {
				r.Get("/lookup/{steamid:[0-9]+}/{slot:[0-9]+}", con.LookUpCharacterID)
				r.Get("/deleted/{steamid:[0-9]+}", con.GetDeletedCharacters)
				r.Get("/{steamid:[0-9]+}", con.GetCharacters)
				r.Get("/{uuid}", con.GetCharacterByIDExternal)
				r.Patch("/restore/{uuid}", con.RestoreCharacter)
				r.Get("/export/{uuid}", con.ExportCharacter)
			})

			r.Route("/rollback/character", func(r chi.Router) {
				r.Get("/{uuid}", con.GetCharacterVersions)
				r.Patch("/{uuid}/latest", con.RollbackCharToLatest)
				r.Patch("/{uuid}/{version:[0-9]+}", con.RollbackCharToVersion)
				r.Delete("/{uuid}", con.DeleteCharRollbacks)
			})

			r.Route("/unsafe/character", func(r chi.Router) {
				r.Patch("/move/{uuid}/to/{steamid:[0-9]+}/{slot:[0-9]+}", con.UnsafeMoveCharacter)
				r.Patch("/copy/{uuid}/to/{steamid:[0-9]+}/{slot:[0-9]+}", con.UnsafeCopyCharacter)
				r.Delete("/delete/{uuid}", con.UnsafeDeleteCharacter)
			})

			r.Route("/user", func(r chi.Router) {
				r.Get("/{steamid:[0-9]+}", con.GetUser)

				r.Patch("/ban/{steamid:[0-9]+}", con.PatchBanSteamID)
				r.Patch("/unban/{steamid:[0-9]+}", con.PatchUnBanSteamID)
				r.Patch("/admin/{steamid:[0-9]+}", con.PatchAdminSteamID)
				r.Patch("/unadmin/{steamid:[0-9]+}", con.PatchUnAdminSteamID)
				r.Patch("/donor/{steamid:[0-9]+}", con.PatchDonorSteamID)
				r.Patch("/undonor/{steamid:[0-9]+}", con.PatchUnDonorSteamID)

				//r.Get("/isdonor/{steamid:[0-9]+}", con.GetIsDonorSteamID)

				if flags.debug {
					r.Get("/list", con.GetAllUsers)
				}
			})

			r.Get("/refresh", func(w http.ResponseWriter, r *http.Request) {
				if err := a.LoadLists(); err != nil {
					response.Error(w, err)
					return
				}

				response.OK(w, true)
			})
		})

		if flags.debug {
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				response.OK(w, true)
			})

			r.Mount("/debug", cmw.Profiler())
		}
	})

	// Let the game server know that's it's trying to use the old API.
	router.Route(static.OldAPIVersion, func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			if !flags.debug {
				r.Use(mw.InternalAuth)
			}

			r.Get("/sc/{hash}", con.DepreciatedAPIVersion)
			r.Get("/ban/{steamid:[0-9]+}", con.DepreciatedAPIVersion)
			r.Get("/map/{name}/{hash}", con.DepreciatedAPIVersion)
			r.Get("/ping", con.DepreciatedAPIVersion)

			r.Route("/character", func(r chi.Router) {
				r.Get("/{steamid:[0-9]+}/{slot:[0-9]}", con.DepreciatedAPIVersion)
				r.Get("/export/{steamid:[0-9]+}/{slot:[0-9]}", con.DepreciatedAPIVersion)

				r.Post("/", con.DepreciatedAPIVersion)
				r.Put("/{uid}", con.DepreciatedAPIVersion)
				r.Delete("/{uid}", con.DepreciatedAPIVersion)
			})
		})
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

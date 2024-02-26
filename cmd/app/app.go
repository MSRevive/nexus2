package app

import (
	"fmt"
	"os"
	"context"
	"io"
	"time"
	"net/http"
	"log/slog"
	"strconv"
	"crypto/tls"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/static"
	"github.com/msrevive/nexus2/pkg/loghandler"

	"github.com/saintwish/kv/ccmap"
	rw "github.com/saintwish/rotatewriter"
	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"
)

type App struct {
	Config *config.Config
	DB database.Database
	Logger *slog.Logger
	List struct {
		SystemAdmin *ccmap.Cache[string, string]
		IP *ccmap.Cache[string, string]
		Ban *ccmap.Cache[string, bool]
		Map *ccmap.Cache[string, uint32]
		Admin *ccmap.Cache[string, bool]
	}

	httpServer *http.Server
}

func New() (app *App) {
	app = &App{}
	app.List.IP = ccmap.New[string, string]()
	app.List.SystemAdmin = ccmap.New[string, string]()
	app.List.Ban = ccmap.New[string, bool]()
	app.List.Map = ccmap.New[string, uint32]()
	app.List.Admin = ccmap.New[string, bool]()

	return
}

func (a *App) LoadConfig(path string) (err error) {
	a.Config, err = config.Load(path)

	return
}

func (a *App) InitializeLogger() {
	iow := io.MultiWriter(os.Stdout, &rw.RotateWriter{
		Dir: a.Config.Log.Dir,
		Filename: "server.log",
		ExpireTime: a.Config.Log.ExpireTime,
		MaxSize: 5 * rw.Megabyte,
	})

	slevel := slog.LevelInfo
	switch(a.Config.Log.Level) {
	case "info":
		slevel = slog.LevelInfo
	case "warn":
		slevel = slog.LevelWarn
	case "error": 
		slevel = slog.LevelError
	case "debug":
		slevel = slog.LevelDebug
	}

	a.Logger = slog.New(loghandler.New(iow, &loghandler.Options{
		Level: slevel,
	}))
}

func (a *App) LoadLists() error {
	if err := a.loadSystemAdminList(a.Config.ApiAuth.SystemAdmins); err != nil {
		return fmt.Errorf("failed to load system admin list: %w", err)
	}

	if a.Config.ApiAuth.EnforceIP {
		if err := a.loadIPList(a.Config.ApiAuth.IPListFile); err != nil {
			return fmt.Errorf("failed to load IP whitelist: %w", err)
		}
	}

	if a.Config.Verify.EnforceMap {
		if err := a.loadMapList(a.Config.Verify.MapListFile); err != nil {
			return fmt.Errorf("failed to load map list: %w", err)
		}
	}

	if a.Config.Verify.EnforceBan {
		if err := a.loadBanList(a.Config.Verify.BanListFile); err != nil {
			return fmt.Errorf("failed to load ban list: %w", err)
		}
	}

	if err := a.loadAdminList(a.Config.Verify.AdminListFile); err != nil {
		return fmt.Errorf("failed to load admin list: %w", err)
	}

	return nil
}

func (a *App) loadIPList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.IP.LoadFromJSON(file)
}

func (a *App) loadMapList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.Map.LoadFromJSON(file)
}

func (a *App) loadBanList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.Ban.LoadFromJSON(file)
}

func (a *App) loadAdminList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.Admin.LoadFromJSON(file)
}

func (a *App) loadSystemAdminList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.SystemAdmin.LoadFromJSON(file)
}

func (a *App) Start(mux chi.Router) error {
	a.Logger.Info("Starting Nexus2", "App Version", static.Version, "Go Version", static.GoVersion, "OS", static.OS, "Arch", static.OSArch)

	a.Logger.Info("Connecting to database")
	if err := a.DB.Connect(a.Config.Database); err != nil {
		return err
	}

	a.httpServer = &http.Server{
		Handler:      mux,
		Addr:         a.Config.Core.Address + ":" + strconv.Itoa(a.Config.Core.Port),
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
	}

	cron := cron.New()
	cron.AddFunc("*/30 * * * *", func(){
		a.Logger.Info("Saving characters from database cache...")
		t1 := time.Now()
		if err := a.DB.SaveToDatabase(); err != nil {
			a.Logger.Error("Failed to save characters!", "error", err)
			return
		}
		a.DB.ClearCache()
		a.Logger.Info("Finished saving to database.", "ping", time.Since(t1))
	})
	cron.Start()

	if a.Config.Cert.Enable {
		a.Logger.Info("Starting HTTPS server with cert")
		return a.StartHTTPWithCert()
	}else{
		a.Logger.Info("Starting HTTP server")
		return a.StartHTTP()
	}

	return nil
}

func (a *App) Close() error {
	a.Logger.Info("Saving characters from database cache...")
	t1 := time.Now()
	if err := a.DB.SaveToDatabase(); err != nil {
		return err
	}
	a.DB.ClearCache()
	a.Logger.Info("Finished saving to database.", "ping", time.Since(t1))

	//close database connection
	a.Logger.Info("Closing database connection")
	if err := a.DB.Disconnect(); err != nil {
		return err
	}

	//try to gracefully shutdown http server with 5 second timeout.
	a.Logger.Info("Shutting down HTTP server gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err // failure/timeout shutting down the server gracefully
	}

	return nil
}
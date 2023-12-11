package app

import (
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
)

type App struct {
	Config *config.Config
	DB database.Database
	Logger *slog.Logger
	List struct {
		IP *ccmap.Cache[string, string]
		Ban *ccmap.Cache[string, bool]
		Map *ccmap.Cache[string, uint32]
		Admin *ccmap.Cache[string, bool]
	}

	httpServer *http.Server
}

func New(cfg *config.Config, db database.Database) (app *App) {
	app = &App{}
	app.Config = cfg
	app.DB = db
	app.List.IP = ccmap.New[string, string]()
	app.List.Ban = ccmap.New[string, bool]()
	app.List.Map = ccmap.New[string, uint32]()
	app.List.Admin = ccmap.New[string, bool]()

	return
}

func (a *App) InitializeLoggers() error {
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

	return nil
}

func (a *App) LoadIPList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.IP.LoadFromJSON(file)
}

func (a *App) LoadMapList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.Map.LoadFromJSON(file)
}

func (a *App) LoadBanList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.Ban.LoadFromJSON(file)
}

func (a *App) LoadAdminList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return a.List.Admin.LoadFromJSON(file)
}

func (a *App) Start(mux chi.Router) error {
	a.Logger.Info("Starting Nexus2", "App Version", static.Version, "Go Version", static.GoVersion, "OS", static.OS, "Arch", static.OSArch)

	a.Logger.Info("Connecting to database")
	if err := a.DB.Connect(a.Config.Database.Connection); err != nil {
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
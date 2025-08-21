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
	"github.com/msrevive/nexus2/pkg/utils"

	"github.com/saintwish/kv/ccmap"
	rw "github.com/saintwish/rotatewriter"
	"github.com/go-chi/chi/v5"
)

type App struct {
	Config *config.Config
	DB database.Database
	Logger *slog.Logger
	List struct {
		SystemAdmin *ccmap.Cache[string, string]
		IP *ccmap.Cache[string, string]
		Map *ccmap.Cache[string, uint32]
	}

	httpServer *http.Server
}

func New() (app *App) {
	app = &App{}
	app.List.SystemAdmin = ccmap.New[string, string]()
	app.List.IP = ccmap.New[string, string]()
	app.List.Map = ccmap.New[string, uint32]()

	return
}

/*
func (a *App) CalcHashes() error {
	if (!a.Config.Verify.EnforceBins) {
		return nil
	}

	// Calculate hash for win32 server binary
	a.Logger.Info("Calculating hash for Server win32 binary", "file", a.Config.Verify.ServerWinBin)
	fh, err := os.Open(a.Config.Verify.ServerWinBin)
	if err != nil {
		return fmt.Errorf("unable to open server win32 binary: %v\n", err)
	}

	hasher := crc32.NewIEEE()
	if _, err := io.Copy(hasher, fh); err != nil {
		return fmt.Errorf("unable to hash: %v\n", err)
	}
	a.Hashes.ServerWin = hasher.Sum32()

	// Calculate hash for unix server binary
	a.Logger.Info("Calculating hash for Server unix binary", "file", a.Config.Verify.ServerUnixBin)
	fh, err = os.Open(a.Config.Verify.ServerUnixBin)
	if err != nil {
		return fmt.Errorf("unable to open server unix binary: %v\n", err)
	}

	hasher = crc32.NewIEEE()
	if _, err := io.Copy(hasher, fh); err != nil {
		return fmt.Errorf("unable to hash: %v\n", err)
	}
	a.Hashes.ServerUnix = hasher.Sum32()

	// Calculate hash for scripts binary
	a.Logger.Info("Calculating hash for scripts binary", "file", a.Config.Verify.ScriptsBin)
	fh, err = os.Open(a.Config.Verify.ScriptsBin)
	if err != nil {
		return fmt.Errorf("unable to open scripts binary: %v\n", err)
	}

	hasher = crc32.NewIEEE()
	if _, err := io.Copy(hasher, fh); err != nil {
		return fmt.Errorf("unable to hash: %v\n", err)
	}
	a.Hashes.Scripts = hasher.Sum32()

	return nil
}*/

func (a *App) LoadConfig(path string) (err error) {
	a.Config, err = config.Load(path)

	return
}

func (a *App) InitializeLogger() (err error) {
	err = os.MkdirAll(a.Config.Log.Dir, os.ModePerm)

	logExpire, err := utils.ParseDuration(a.Config.Log.ExpireTime)
	if err != nil {
		return
	}

	iow := io.MultiWriter(os.Stdout, &rw.RotateWriter{
		Dir: a.Config.Log.Dir,
		Filename: "server.log",
		ExpireTime: logExpire,
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

	return
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
	if err := a.DB.Connect(a.Config.Database, database.Options{
		Logger: a.SetUpDatabaseLogger(),
	}); err != nil {
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
package app

import (
	"os"
	"context"
	"io"
	"time"
	"net/http"
	"log/slog"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/config"

	"github.com/saintwish/kv/ccmap"
	rw "github.com/saintwish/rotatewriter"
)

type App struct {
	Config config.Config
	DB database.Database
	HTTPServer *http.Server
	Logger struct {
		Core *slog.Logger
		API *slog.Logger
	}
	List struct {
		IP *ccmap.Cache[string, string]
		Ban *ccmap.Cache[string, bool]
		Map *ccmap.Cache[string, uint32]
		Admin *ccmap.Cache[string, bool]
	}
}

func New(cfg config.Config, db database.Database) (app *App) {
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
	expiration, err := time.ParseDuration(a.Config.Log.ExpireTime)
	if err != nil {
		return err
	}

	iow := io.MultiWriter(os.Stdout, &rw.RotateWriter{
		Dir: a.Config.Log.Dir,
		Filename: "server.log",
		ExpireTime: expiration,
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

	a.Logger.Core = slog.New(NewLogHandler(iow, &LogOptions{
		Level: slevel,
		Domain: "CORE",
	}))
	a.Logger.API = slog.New(NewLogHandler(iow, &LogOptions{
		Level: slevel,
		Domain: "API",
	}))

	return nil
}

func (a *App) SetHTTPServer(srv *http.Server) {
	a.HTTPServer = srv
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

func (a *App) Start() error {
	a.Logger.Core.Info("Connecting to database")
	if err := a.DB.Connect(a.Config.Database.Connection); err != nil {
		return err
	}

	if a.Config.Cert.Enable {
		a.Logger.Core.Info("Starting HTTPS server with cert")
		return a.StartHTTPWithCert()
	}else{
		a.Logger.Core.Info("Starting HTTP server")
		return a.StartHTTP()
	}

	return nil
}

func (a *App) Close() error {
	//close database connection
	a.Logger.Core.Info("Closing database connection")
	if err := a.DB.Disconnect(); err != nil {
		return err
	}

	//try to gracefully shutdown http server with 5 second timeout.
	a.Logger.Core.Info("Shutting down HTTP server gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		return err // failure/timeout shutting down the server gracefully
	}

	return nil
}
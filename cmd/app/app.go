package app

import (
	"os"
	"sync"
	"context"
	"errors"
	"fmt"
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
	Config Config
	DB database.Database
	HTTPServer *http.Server
	Logger struct {
		Core *slog.Logger
		API *slog.Logger
	}
	List struct {
		IP ccmap.Cache[string, string]
		Ban ccmap.Cache[string, bool]
		Map ccmap.Cache[string, uint32]
		Admin ccmap.Cache[string, bool]
	}
}

func New(cfg config.Config, db database.Database) (app *App) {
	app = &App{}
	app.Config = cfg
	app.DB = db
	app.List.IPList = ccmap.New[string, string]()
	app.List.BanList = ccmap.New[string, bool]()
	app.List.MapList = ccmap.New[string, uint32]()
	app.List.AdminList = ccmap.New[string, bool]()

	return
}

func (a *App) InitializeLoggers() {
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
}

func (a *App) SetHTTPServer(srv *http.Server) {
	a.HTTPServer = srv
}

func (a *App) LoadIPList(path string) (err error) {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = a.List.IP.LoadFromJSON(file)

	return
}

func (a *App) LoadMapList(path string) (err error) {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = a.List.Map.LoadFromJSON(file)

	return
}

func (a *App) LoadBanList(path string) (err error) {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = a.List.Ban.LoadFromJSON(file)

	return
}

func (a *App) LoadAdminList(path string) (err error) {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = a.List.Admin.LoadFromJSON(file)

	return
}

func (a *App) Start() error {
	fmt.Println("Connecting to database.")
	if err := a.DB.Connect(); err != nil {
		return err
	}

	if a.Config.Cert.Enable {
		return a.StartHTTPWithCert()
	}else{
		return a.StartHTTP()
	}

	return nil
}

func (a *App) Close() error {
	//close database connection
	fmt.Println("Disconnecting from database.")
	if err := a.DB.Disconnect(); err != nil {
		return err
	}

	//try to gracefully shutdown http server with 5 second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		return err // failure/timeout shutting down the server gracefully
	}

	return nil
}
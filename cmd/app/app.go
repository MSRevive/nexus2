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
	
	"github.com/msrevive/nexus2/internal/database"

	"github.com/saintwish/auralog"
	"github.com/saintwish/kv/ccmap"
)

type App struct {
	Config Config
	DB database.Database
	HTTPServer *http.Server
	Logger struct {
		Core *auralog.Logger
		API *auralog.Logger
	}
	List struct {
		IPList ccmap.Cache[string, string]
		BanList ccmap.Cache[string, bool]
		MapList ccmap.Cache[string, uint32]
		AdminList ccmap.Cache[string, bool]
	}
}

func New(cfg Config) (app *App, db database.Database) {
	app = &App{}
	app.Config = cfg
	app.DB = db
	app.List.IPList = ccmap.New[string, string]()
	app.List.BanList = ccmap.New[string, bool]()
	app.List.MapList = ccmap.New[string, uint32]()
	app.List.AdminList = ccmap.New[string, bool]()

	return
}

func (a *App) SetupLoggers(logcore *auralog.Logger, logapi *auralog.Logger) {
	a.Logger.Core = logcore
	a.Logger.API = logapi
}

func (a *App) SetHTTPServer(srv *http.Server) {
	a.HTTPServer = srv
}

func (a *App) LoadIPList(path string) (err error) {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = a.List.IPList.LoadFromJSON(file)

	return
}

func (a *App) LoadMapList(path string) (err error) {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = a.List.MapList.LoadFromJSON(file)

	return
}

func (a *App) LoadBanList(path string) (err error) {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = a.List.BanList.LoadFromJSON(file)

	return
}

func (a *App) LoadAdminList(path string) (err error) {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = a.List.AdminList.LoadFromJSON(file)

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
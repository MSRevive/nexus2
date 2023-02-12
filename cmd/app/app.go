package app

import (
	"os"
	"sync"
	
	"github.com/msrevive/nexus2/ent"
	"github.com/saintwish/auralog"

	"github.com/goccy/go-json"
)

var (
	Version = "canary"

	iPListMutex = new(sync.RWMutex)
	banListMutex = new(sync.RWMutex)
	mapListMutex = new(sync.RWMutex)
	adminListMutex = new(sync.RWMutex)
)

type App struct {
	Config *config
	Client *ent.Client
	LogCore *auralog.Logger
	LogAPI *auralog.Logger

	IPList map[string]bool
	BanList map[string]bool
	MapList map[string]uint32
	AdminList map[string]bool
}

func New(cfg *config) *App {
	return &App {
		Config: cfg,
	}
}

func (a *App) SetupLoggers(logcore *auralog.Logger, logapi *auralog.Logger) {
	a.LogCore = logcore
	a.LogAPI = logapi
}

func (a *App) SetupClient(client *ent.Client) {
	a.Client = client
}

func (a *App) LoadIPList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	iPListMutex.Lock()
	_ = json.Unmarshal([]byte(file), &a.IPList)
	iPListMutex.Unlock()

	return nil
}

func (a *App) LoadMapList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	mapListMutex.Lock()
	_ = json.Unmarshal([]byte(file), &a.MapList)
	mapListMutex.Unlock()

	return nil
}

func (a *App) LoadBanList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	banListMutex.Lock()
	_ = json.Unmarshal([]byte(file), &a.BanList)
	banListMutex.Unlock()

	return nil
}

func (a *App) LoadAdminList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	adminListMutex.Lock()
	_ = json.Unmarshal([]byte(file), &a.AdminList)
	adminListMutex.Unlock()

	return nil
}
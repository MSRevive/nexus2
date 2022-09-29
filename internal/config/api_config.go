package config

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

type iCfgAuth interface {
	IsEnforcingIP() bool
	IsKnownIP(ip string) bool
	IsEnforcingKey() bool
	IsValidKey(key string) bool
	GetUserAgent() string
}

type iCfgVerify interface {
	VerifyMapName(name string, calculated uint32) bool
	VerifySC(calculated uint32) bool
	IsSteamIDBanned(steamid string) bool
	IsSteamIDAdmin(steamid string) bool
	EnforceAndVerifyBanned(steamid string) bool
}

type iCfgHelper interface {
	GetDBString() string
	GetMaxRequests() int
	GetMaxAge() time.Duration
	GetDebugMode() bool
}

// Verify that apiConfig implements these interfaces
var (
	_ iCfgAuth   = (*apiConfig)(nil)
	_ iCfgVerify = (*apiConfig)(nil)
	_ iCfgHelper = (*apiConfig)(nil)
)

type apiConfig struct {
	Core struct {
		MaxThreads int
		Graceful   time.Duration
		RootPath   string
		DBString   string
		DebugMode  bool
	}
	RateLimit struct {
		Enable      bool
		MaxRequests int
		MaxAge      time.Duration
	}
	Cert struct {
		Enable bool
		Domain string
	}
	ApiAuth struct {
		EnforceKey bool
		EnforceIP  bool
		Key        string
		IPListFile string
	}
	Verify struct {
		EnforceBan    bool
		EnforceMap    bool
		EnforceSC     bool
		MapListFile   string
		BanListFile   string
		AdminListFile string
		SCHash        uint32
		Useragent     string
	}
	Log struct {
		Level      string
		Dir        string
		ExpireTime string
	}

	iPList         map[string]bool
	iPListMutex    sync.RWMutex
	banList        map[string]bool
	banListMutex   sync.RWMutex
	mapList        map[string]uint32
	mapListMutex   sync.RWMutex
	adminList      map[string]bool
	adminListMutex sync.RWMutex
}

func LoadConfig(path string, dbg bool) (*apiConfig, error) {
	var cfg apiConfig

	switch filepath.Ext(path) {
	case ".toml", ".ini":
		if err := ini.MapTo(&cfg, path); err != nil {
			return nil, err
		}

	case ".yaml", ".json", ".yml":
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("unsupported config type")
	}

	cfg.Core.DebugMode = dbg
	return &cfg, nil
}

func (cfg *apiConfig) LoadIPList() error {
	cfg.iPListMutex.Lock()
	defer cfg.iPListMutex.Unlock()

	return loadJsonFile(cfg.ApiAuth.IPListFile, &cfg.iPList)
}

func (cfg *apiConfig) LoadMapList() error {
	cfg.mapListMutex.Lock()
	defer cfg.mapListMutex.Unlock()

	return loadJsonFile(cfg.Verify.MapListFile, &cfg.mapList)
}

func (cfg *apiConfig) LoadBanList() error {
	cfg.banListMutex.Lock()
	defer cfg.banListMutex.Unlock()

	return loadJsonFile(cfg.Verify.BanListFile, &cfg.banList)
}

func (cfg *apiConfig) LoadAdminList() error {
	cfg.adminListMutex.Lock()
	defer cfg.adminListMutex.Unlock()

	return loadJsonFile(cfg.Verify.AdminListFile, &cfg.adminList)
}

func (cfg *apiConfig) Migrate() error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile("./runtime/config.yaml", data, 0655); err != nil {
		return err
	}

	return nil
}

func (c *apiConfig) IsEnforcingIP() bool {
	return c.ApiAuth.EnforceIP
}

func (c *apiConfig) IsKnownIP(ip string) bool {
	c.iPListMutex.RLock()
	defer c.iPListMutex.RUnlock()

	_, ok := c.iPList[ip]
	return ok
}

func (c *apiConfig) IsEnforcingKey() bool {
	return c.ApiAuth.EnforceKey
}

func (c *apiConfig) IsValidKey(key string) bool {
	return c.ApiAuth.Key == key
}

func (c *apiConfig) GetUserAgent() string {
	return c.Verify.Useragent
}

func (c *apiConfig) VerifyMapName(name string, calculated uint32) bool {
	if c.Verify.EnforceMap {
		c.mapListMutex.RLock()
		defer c.mapListMutex.RUnlock()

		return c.mapList[name] == calculated
	}

	return true // return true if enforcemap is false
}

func (c *apiConfig) VerifySC(calculated uint32) bool {
	if c.Verify.EnforceSC {
		return c.Verify.SCHash == calculated
	}

	return true // return true if enforcesc is false
}

func (c *apiConfig) IsSteamIDAdmin(steamid string) bool {
	c.adminListMutex.RLock()
	defer c.adminListMutex.RUnlock()

	_, ok := c.adminList[steamid]
	return ok
}

func (c *apiConfig) IsSteamIDBanned(steamid string) bool {
	c.banListMutex.RLock()
	defer c.banListMutex.RUnlock()

	_, ok := c.banList[steamid]
	return ok
}

func (c *apiConfig) EnforceAndVerifyBanned(steamid string) bool {
	if c.Verify.EnforceBan {
		return c.IsSteamIDBanned(steamid)
	}
	return false
}

func (c *apiConfig) GetDBString() string {
	return c.Core.DBString
}

func (c *apiConfig) GetMaxRequests() int {
	return c.RateLimit.MaxRequests
}

func (c *apiConfig) GetMaxAge() time.Duration {
	return c.RateLimit.MaxAge
}

func (c *apiConfig) GetDebugMode() bool {
	return c.Core.DebugMode
}

func loadJsonFile(path string, container interface{}) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(file), container)
}

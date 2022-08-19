package system

import(
	"os"
	"time"
	"sync"
	"errors"
	"path/filepath"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
	"github.com/goccy/go-json"
)

var (
	Version = "canary"
	AuthCfg iCfgAuth = (*config)(nil)
	VerifyCfg iCfgVerify = (*config)(nil)
	HelperCfg iCfgHelper = (*config)(nil)
)

type config struct {
	Core struct {
		MaxThreads int
		Graceful time.Duration
		RootPath string
		DBString string
		DebugMode bool
	}
	RateLimit struct {
		Enable bool
		MaxRequests int
		MaxAge time.Duration
	}
	Cert struct {
		Enable bool
		Domain string
	}
	ApiAuth struct {
		EnforceKey bool
		EnforceIP bool
		Key string
		IPListFile string
	}
	Verify struct {
		EnforceBan bool
		EnforceMap bool
		EnforceSC bool
		MapListFile string
		BanListFile string
		AdminListFile string
		SCHash uint32
		Useragent string
	}
	Log struct {
		Level string
		Dir string
		ExpireTime string
	}

	iPList map[string]bool
	iPListMutex sync.RWMutex
	banList map[string]bool
	banListMutex sync.RWMutex
	mapList map[string]uint32
	mapListMutex sync.RWMutex
	adminList map[string]bool
	adminListMutex sync.RWMutex
}

func LoadConfig(path string, dbg bool) (*config, error) {
	var cfg config

	switch filepath.Ext(path) {
	case ".toml", ".ini":
		if err := ini.MapTo(&cfg, path); err != nil {
			return nil, err
		}

	case ".yaml", ".json", ".yml":
		data,err := os.ReadFile(path)
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

func (cfg *config) LoadIPList() error {
	cfg.iPListMutex.Lock()
	defer cfg.iPListMutex.Unlock()

	return loadJsonFile(cfg.ApiAuth.IPListFile, &cfg.iPList)
}

func (cfg *config) LoadMapList() error {
	cfg.mapListMutex.Lock()
	defer cfg.mapListMutex.Unlock()

	return loadJsonFile(cfg.Verify.MapListFile, &cfg.mapList)
}

func (cfg *config) LoadBanList() error {
	cfg.banListMutex.Lock()
	defer cfg.banListMutex.Unlock()

	return loadJsonFile(cfg.Verify.BanListFile, &cfg.banList)
}

func (cfg *config) LoadAdminList() error {
	cfg.adminListMutex.Lock()
	defer cfg.adminListMutex.Unlock()

	return loadJsonFile(cfg.Verify.AdminListFile, &cfg.adminList)
}

func (cfg *config) Migrate() error {
	data,err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile("./runtime/config.yaml", data, 0655); err != nil {
		return err
	}
	
	return nil
}

func loadJsonFile(path string, container interface{}) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(file), container)
}
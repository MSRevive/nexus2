package app

import (
	"os"
	"time"
	"errors"
	"path/filepath"
  
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Core struct {
		Address string
		Port int
		MaxThreads int
		Graceful time.Duration
		RootPath string
		DBString string
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
	Char struct {
		MaxBackups int
		BackupTime string
	}
	Log struct {
		Level string
		Dir string
		ExpireTime string
	}
}
  
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, os.ErrNotExist
	}

	switch filepath.Ext(path) {
	case ".toml", ".ini":
		if err := ini.MapTo(&cfg, path); err != nil {
			return nil, err
		}
	case ".yaml", ".json", ".yml":
		data,err := os.ReadFile(path)
		if data != nil {
			err = yaml.Unmarshal(data, &cfg)
		}

		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported config type")
	}

	return &cfg, nil
}

func (a *App) MigrateConfig() error {
	data,err := yaml.Marshal(a.Config)
	if err != nil {
		return err
	}

	if err := os.WriteFile("./runtime/config.yaml", data, 0655); err != nil {
		return err
	}

	return nil
}
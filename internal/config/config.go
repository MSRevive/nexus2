package config

import (
	"os"
	"fmt"
	"errors"
	"time"
	"path/filepath"
  
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Core struct {
		Address string
		Port int
		Timeout time.Duration
	}
	Database struct {
		Connection string
	}
	RateLimit struct {
		MaxRequests int
		MaxAge string
	}
	Cert struct {
		Enable bool
		Domain string
	}
	ApiAuth struct {
		EnforceKey bool
		EnforceIP bool
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
	case ".yaml", ".json", ".yml":
		data,err := os.ReadFile(path)
		if data != nil {
			err = yaml.Unmarshal(data, &cfg)
		}

		if err != nil {
			return nil, err
		}
		return &cfg, nil

	case ".toml", ".ini":
		if err := ini.MapTo(&cfg, path); err != nil {
			return nil, err
		}
		return &cfg, nil
		
	default:
		return nil, fmt.Errorf("%s", "unsupported config type")
	}

	return nil, fmt.Errorf("Failed to read config file: %s", path)
}
package config

import (
	"os"
	"fmt"
	"errors"
	"path/filepath"
  
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Core struct {
		Address string
		Port int
	}
	Database struct {
		Conn string
		MaxIdleConns int
		MaxOpenConns int
		ConnMaxLifetime string
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
  
func LoadConfig(path string) (Config, error) {
	var cfg Config

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return cfg, os.ErrNotExist
	}

	switch filepath.Ext(path) {
	case ".yaml", ".json", ".yml":
		data,err := os.ReadFile(path)
		if data != nil {
			err = yaml.Unmarshal(data, &cfg)
		}

		if err != nil {
			return cfg, err
		}
		return cfg, nil

	case ".toml", ".ini":
		if err := ini.MapTo(&cfg, path); err != nil {
			return cfg, err
		}
		return cfg, nil
		
	default:
		return cfg, fmt.Errorf("%s", "unsupported config type")
	}

	return cfg, fmt.Errorf("Failed to read config file: %s", path)
}
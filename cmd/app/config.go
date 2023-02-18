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
	Database struct {
		Conn string
		MaxIdleConns int
		MaxOpenConns int
		ConnMaxLifetime string
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
		return nil, errors.New("unsupported config type")
	}

	return nil, errors.New(fmt.Sprintf("Failed to read config file: %s", path))
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
package config

import (
	"os"
	"fmt"
	"time"
	"path/filepath"

	"github.com/msrevive/nexus2/internal/database"
  
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Core struct {
		Address string
		Port int
		Timeout time.Duration
		DBType string
	}
	Database database.Config
	RateLimit struct {
		MaxRequests int
		MaxAge string
	}
	Cert struct {
		Enable bool
		Domain string
	}
	ApiAuth struct {
		SystemAdmins string
		EnforceKey bool
		EnforceIP bool
		IPListFile string
		UserAgent string
	}
	Verify struct {
		EnforceBan bool
		EnforceMap bool
		EnforceBins bool
		MapListFile string
		BanListFile string
		AdminListFile string
		ServerUnixBin string
		ServerWinBin string
		ScriptsBin string
	}
	Char struct {
		MaxBackups int
		BackupTime time.Duration
		DeletedExpireTime time.Duration
	}
	Log struct {
		Level string
		Dir string
		ExpireTime time.Duration
	}
}
  
func Load(path string) (cfg *Config, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	switch filepath.Ext(path) {
	case ".yaml", ".json", ".yml":
		err = yaml.Unmarshal(data, &cfg)
	case ".toml", ".ini":
		err = ini.MapTo(&cfg, path)
	default:
		err = fmt.Errorf("%s", "unsupported config type")
	}

	return
}
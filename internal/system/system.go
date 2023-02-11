package system

import(
  "os"
  "time"
  "sync"
  "errors"
  "path/filepath"

  "github.com/msrevive/nexus2/ent"

  "gopkg.in/ini.v1"
  "gopkg.in/yaml.v2"
  "github.com/goccy/go-json"
)

var (
  Client *ent.Client
  Config config
  Dbg bool
  
  IPList map[string]bool
  iPListMutex = new(sync.RWMutex)
  BanList map[string]bool
  banListMutex = new(sync.RWMutex)
  MapList map[string]uint32
  mapListMutex = new(sync.RWMutex)
  AdminList map[string]bool
  adminListMutex = new(sync.RWMutex)
)

type config struct {
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
  Log struct {
    Level string
    Dir string
    ExpireTime string
  }
}

func LoadConfig(path string) error {
  switch filepath.Ext(path) {
  case ".toml", ".ini":
    if err := ini.MapTo(&Config, path); err != nil {
      return err
    }
    
    return nil
  case ".yaml", ".json", ".yml":
    data,err := os.ReadFile(path)
    if data != nil {
      err = yaml.Unmarshal(data, &Config)
    }
    
    if err != nil {
      return err
    }
    
    return nil
  default:
    return errors.New("unsupported config type")
  }
}

func MigrateConfig() error {
  data,err := yaml.Marshal(Config)
  if err != nil {
    return err
  }
  
  if err := os.WriteFile("./runtime/config.yaml", data, 0655); err != nil {
    return err
  }
  
  return nil
}

func LoadIPList(path string) error {
  file,err := os.ReadFile(path)
  if err != nil {
    return err
  }
  
  iPListMutex.Lock()
  _ = json.Unmarshal([]byte(file), &IPList)
  iPListMutex.Unlock()
  
  return nil
}

func LoadMapList(path string) error {
  file,err := os.ReadFile(path)
  if err != nil {
    return err
  }
  
  mapListMutex.Lock()
  _ = json.Unmarshal([]byte(file), &MapList)
  mapListMutex.Unlock()
  
  return nil
}

func LoadBanList(path string) error {
  file,err := os.ReadFile(path)
  if err != nil {
    return err
  }
  
  banListMutex.Lock()
  _ = json.Unmarshal([]byte(file), &BanList)
  banListMutex.Unlock()
  
  return nil
}

func LoadAdminList(path string) error {
  file,err := os.ReadFile(path)
  if err != nil {
    return err
  }
  
  adminListMutex.Lock()
  _ = json.Unmarshal([]byte(file), &AdminList)
  adminListMutex.Unlock()
  
  return nil
}

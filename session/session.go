package session

import(
  "time"
  "sync"
  "encoding/json"
  "io/ioutil"

  "github.com/msrevive/nexus2/ent"

  "gopkg.in/ini.v1"
)

var (
  Client *ent.Client
  Config config
  Dbg bool
  
  IPList map[string]bool
  IPListMutex = new(sync.RWMutex)
  BanList map[string]bool
  BanListMutex = new(sync.RWMutex)
  MapList map[string]uint32
  MapListMutex = new(sync.RWMutex)
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
    SCHash uint32
  }
  Log struct {
    Level string
    Dir string
    ExpireTime string
  }
}

func LoadConfig(path string) error {
  if err := ini.MapTo(&Config, path); err != nil {
    return err
  }
  
  return nil
}

func LoadIPList(path string) error {
  file,err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }
  
  IPListMutex.Lock()
  _ = json.Unmarshal([]byte(file), &IPList)
  IPListMutex.Unlock()
  
  return nil
}

func LoadMapList(path string) error {
  file,err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }
  
  MapListMutex.Lock()
  _ = json.Unmarshal([]byte(file), &MapList)
  MapListMutex.Unlock()
  
  return nil
}

func LoadBanList(path string) error {
  file,err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }
  
  BanListMutex.Lock()
  _ = json.Unmarshal([]byte(file), &BanList)
  BanListMutex.Unlock()
  
  return nil
}

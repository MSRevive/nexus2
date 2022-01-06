package session

import(
  "time"
  "encoding/json"
  "io/ioutil"

  //"github.com/msrevive/nexus2/ent"

  "github.com/BurntSushi/toml"
)

var (
  //Client *ent.Client
  Config config
  IPList map[string]int8
  BanList map[uint64]int8
  MapList map[uint32]int8
)

type config struct {
  Core struct {
    Port int
    IP string
    MaxThreads int
    MaxRequests int
    MaxAge time.Duration
    Graceful time.Duration
    RootPath string
  }
  RateLimit struct {
    Enable bool
    MaxRequests int
    MaxAge time.Duration
  }
  TLS struct {
    Enable bool
    Port int
    CertFile string
    KeyFile string
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
    MapListFile string
    BanListFile string
    SCHash uint32
  }
  Log struct {
    Dir string
  }
}

func LoadConfig(path string) error {
  _, err := toml.DecodeFile(path, &Config);
  if err != nil {
    return err
  }

  return nil
}

func LoadIPList(path string) error {
  file,err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }
  
  _ = json.Unmarshal([]byte(file), &IPList)
  
  return nil
}

func LoadMapList(path string) error {
  file,err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }
  
  _ = json.Unmarshal([]byte(file), &MapList)
  
  return nil
}

func LoadBanList(path string) error {
  file,err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }
  
  _ = json.Unmarshal([]byte(file), &BanList)
  
  return nil
}

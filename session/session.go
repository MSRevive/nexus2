package session

import(
  "time"

  "github.com/BurntSushi/toml"
)

var (
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

package system

import(
  "time"
)

type iCfgAuth interface {
	IsEnforcingIP() bool
	IsKnownIP(ip string) bool
	IsEnforcingKey() bool
  IsValidKey(key string) bool
  GetUserAgent() string
}

type iCfgVerify interface {
  VerifyMapName(name string, calculated uint32) bool
  VerifySC(calculated uint32) bool
  IsSteamIDBanned(steamid string) bool
  IsSteamIDAdmin(steamid string) bool
  EnforceAndVerifyBanned(steamid string) bool
}

type iCfgHelper interface {
  GetDBString() string
  GetMaxRequests() int
  GetMaxAge() time.Duration
  GetDebugMode() bool
}

func (c *config) IsEnforcingIP() bool {
	return c.ApiAuth.EnforceIP
}

func (c *config) IsKnownIP(ip string) bool {
	c.iPListMutex.RLock()
	defer c.iPListMutex.RUnlock()

	_, ok := c.iPList[ip]
	return ok
}

func (c *config) IsEnforcingKey() bool {
	return c.ApiAuth.EnforceKey
}

func (c *config) IsValidKey(key string) bool {
	return c.ApiAuth.Key == key
}

func (c *config) GetUserAgent() string {
  return c.Verify.Useragent
}

func (c *config) VerifyMapName(name string, calculated uint32) bool {
  if c.Verify.EnforceMap {
    c.mapListMutex.RLock()
    defer c.mapListMutex.RUnlock()

    return c.mapList[name] == calculated
  }

  return true //return true if enforcemap is false
}

func (c *config) VerifySC(calculated uint32) bool {
  if c.Verify.EnforceSC {
    return c.Verify.SCHash == calculated
  }

  return true //return true if enforcesc is false
}

func (c *config) IsSteamIDAdmin(steamid string) bool {
	c.adminListMutex.RLock()
	defer c.adminListMutex.RUnlock()

	_, ok := c.adminList[steamid]
	return ok
}

func (c *config) IsSteamIDBanned(steamid string) bool {
	c.banListMutex.RLock()
	defer c.banListMutex.RUnlock()

	_, ok := c.banList[steamid]
	return ok
}

func (c *config) EnforceAndVerifyBanned(steamid string) bool {
	if c.Verify.EnforceBan {
		return c.IsSteamIDBanned(steamid)
	}
	return false
}

func (c *config) GetDBString() string {
  return c.Core.DBString
}

func (c *config) GetMaxRequests() int {
  return c.RateLimit.MaxRequests
}

func (c *config) GetMaxAge() time.Duration {
  return c.RateLimit.MaxAge
}

func (c *config) GetDebugMode() bool {
  return c.Core.DebugMode
}
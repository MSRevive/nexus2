package controller

import (
  "strconv"
  "net/http"
  
  "github.com/msrevive/nexus2/pkg/response"
  "github.com/msrevive/nexus2/internal/system"
  
  "github.com/gorilla/mux"
)

//GET map/{name}/{hash}
func (c *controller) GetMapVerify(w http.ResponseWriter, r *http.Request) {
  if !system.Config.Verify.EnforceMap {
    response.Result(w, true)
    return
  }
  
  vars := mux.Vars(r)
  name := vars["name"]
  hash, err := strconv.ParseUint(vars["hash"], 10, 32)
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  if res,_ := system.MapList[name]; res == uint32(hash) {
    response.Result(w, true)
    return
  }
  
  response.Result(w, false)
  return
}

//GET ban/{steamid}
//in this case false means player isn't banned
func (c *controller) GetBanVerify(w http.ResponseWriter, r *http.Request) {
  if !system.Config.Verify.EnforceBan {
    response.Result(w, false)
    return
  }
  
  vars := mux.Vars(r)
  steamid := vars["steamid"]
  
  if _,ok := system.BanList[steamid]; ok {
    response.Result(w, true)
    return
  }
  
  response.Result(w, false)
  return
}

//GET sc/{hash}
func (c *controller) GetSCVerify(w http.ResponseWriter, r *http.Request) {
  if !system.Config.Verify.EnforceSC {
    response.Result(w, true)
    return
  }
  
  vars := mux.Vars(r)
  hash, err := strconv.ParseUint(vars["hash"], 10, 32)
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  if system.Config.Verify.SCHash == uint32(hash) {
    response.Result(w, true)
    return
  }
  
  response.Result(w, false)
}

//GET ping
func (c *controller) GetPing(w http.ResponseWriter, r *http.Request) {
  response.Result(w, true)
}
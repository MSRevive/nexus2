package controller

import (
  "strconv"
  "net/http"
  //"encoding/json"
  
  "github.com/msrevive/nexus2/response"
  "github.com/msrevive/nexus2/service"
  "github.com/msrevive/nexus2/system"
  "github.com/msrevive/nexus2/ent"
  "github.com/msrevive/nexus2/log"
  
  "github.com/google/uuid"
  "github.com/gorilla/mux"
  "github.com/goccy/go-json"
)

//GET /character/
func (c *controller) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
  chars, err := service.New(r.Context()).CharactersGetAll()
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  
  response.OK(w, chars)
}

//GET /character/{steamid}
func (c *controller) GetCharacters(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  steamid := vars["steamid"]
  var isBanned bool = false
  var isAdmin bool = false
  
  chars, err := service.New(r.Context()).CharactersGetBySteamid(steamid)
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  
  if _,ok := system.BanList[steamid]; ok && system.Config.Verify.EnforceBan {
    isBanned = true
  }
  
  if _,ok := system.AdminList[steamid]; ok {
    isAdmin = true
  }
  
  response.OKChar(w, isBanned, isAdmin, chars)
}

//GET /character/{steamid}/{slot}
func (c *controller) GetCharacter(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  steamid := vars["steamid"]
  slot, err := strconv.Atoi(vars["slot"])
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  var isBanned bool = false
  var isAdmin bool = false
  
  char, err := service.New(r.Context()).CharacterGetBySteamidSlot(steamid, slot)
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  
  if _,ok := system.BanList[steamid]; ok && system.Config.Verify.EnforceBan {
    isBanned = true
  }
  
  if _,ok := system.AdminList[steamid]; ok {
    isAdmin = true
  }
  
  response.OKChar(w, isBanned, isAdmin, char)
}

//GET /character/id/{uid}
func (c *controller) GetCharacterByID(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  uid, err := uuid.Parse(vars["uid"])
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  var isBanned bool = false
  var isAdmin bool = false
  
  char, err := service.New(r.Context()).CharacterGetByID(uid)
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  
  if _,ok := system.BanList[char.Steamid]; ok && system.Config.Verify.EnforceBan {
    isBanned = true
  }
  
  if _,ok := system.AdminList[char.Steamid]; ok {
    isAdmin = true
  }
  
  response.OKChar(w, isBanned, isAdmin, char)
}

//POST /character/
func (c *controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
  var newChar ent.Character
  err := json.NewDecoder(r.Body).Decode(&newChar)
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  
  char, err := service.New(r.Context()).CharacterCreate(newChar)
  if err != nil {
    log.Log.Errorln(err)
    response.Error(w, err)
    return
  }
  
  response.OK(w, char)
}

//PUT /character/{uid}
func (c *controller) PutCharacter(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  uid, err := uuid.Parse(vars["uid"])
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  
  var updateChar ent.Character
  err = json.NewDecoder(r.Body).Decode(&updateChar)
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  
  char, err := service.New(r.Context()).CharacterUpdate(uid, updateChar)
  if err != nil {
    log.Log.Errorln(err)
    response.Error(w, err)
    return
  }
  
  response.OK(w, char)
}

//DELETE /character/{uid}
func (c *controller) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  uid, err := uuid.Parse(vars["uid"])
  if err != nil {
    log.Log.Errorln(err)
    response.BadRequest(w, err)
    return
  }
  
  err = service.New(r.Context()).CharacterDelete(uid)
  if err != nil {
    log.Log.Errorln(err)
    response.Error(w, err)
    return
  }
  
  response.OK(w, uid)
}
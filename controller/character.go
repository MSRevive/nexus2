package controller

import (
  "strconv"
  "net/http"
  "encoding/json"
  
  "github.com/msrevive/nexus2/response"
  "github.com/msrevive/nexus2/service"
  "github.com/msrevive/nexus2/ent"
  
  "github.com/google/uuid"
  "github.com/gorilla/mux"
)

//GET map/character/
func (c *controller) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
  chars, err := service.New(r.Context()).CharactersGetAll()
  if err != nil {
    response.BadRequest(w, err)
  }
  
  response.OK(w, chars)
}

//GET map/character/{steamid}
func (c *controller) GetCharacters(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  steamid, err := strconv.ParseUint(vars["steamid"], 10, 64)
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  chars, err := service.New(r.Context()).CharactersGetBySteamid(steamid)
  if err != nil {
    response.BadRequest(w, err)
  }
  
  response.OK(w, chars)
}

//GET map/character/{steamid}/{slot}
func (c *controller) GetCharacter(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  steamid, err := strconv.ParseUint(vars["steamid"], 10, 64)
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  slot, err := strconv.Atoi(vars["slot"])
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  chars, err := service.New(r.Context()).CharacterGetBySteamidSlot(steamid, slot)
  if err != nil {
    response.BadRequest(w, err)
  }
  
  response.OK(w, chars)
}

//GET map/character/id/{uid}
func (c *controller) GetCharacterByID(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  uid, err := uuid.Parse(vars["uid"])
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  char, err := service.New(r.Context()).CharacterGetByID(uid)
  if err != nil {
    response.BadRequest(w, err)
  }
  
  response.OK(w, char)
}

//POST map/character/
func (c *controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
  var newChar ent.Character
  err := json.NewDecoder(r.Body).Decode(&newChar)
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  char, err := service.New(r.Context()).CharacterCreate(newChar)
  if err != nil {
    response.Error(w, err)
    return
  }
  
  response.OK(w, char)
}

//PUT map/character/{uid}
func (c *controller) PutCharacter(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  uid, err := uuid.Parse(vars["uid"])
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  var updateChar ent.Character
  err = json.NewDecoder(r.Body).Decode(&updateChar)
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  char, err := service.New(r.Context()).CharacterUpdate(uid, updateChar)
  if err != nil {
    response.Error(w, err)
    return
  }
  
  response.OK(w, char)
}

//DELETE map/character/{uid}
func (c *controller) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  uid, err := uuid.Parse(vars["uid"])
  if err != nil {
    response.BadRequest(w, err)
    return
  }
  
  err = service.New(r.Context()).CharacterDelete(uid)
  if err != nil {
    response.Error(w, err)
    return
  }
  
  response.OK(w, uid)
}
package controller

import (
  "net/http"
  
  "github.com/msrevive/nexus2/response"
  "github.com/msrevive/nexus2/service"
  
  //"github.com/gorilla/mux"
)

//GET map/character/
func (c *controller) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
  chars, err := service.New(r.Context()).CharGetAll()
  if err != nil {
    response.BadRequest(w, err)
  }
  
  response.OK(w, chars)
}

//GET map/character/{steamid}
func (c *controller) GetCharacters(w http.ResponseWriter, r *http.Request) {
  response.Result(w, true)
}

//GET map/character/{steamid}/{slot}
func (c *controller) GetCharacter(w http.ResponseWriter, r *http.Request) {
  response.Result(w, true)
}

//GET map/character/id/{uid}
func (c *controller) GetCharacterByID(w http.ResponseWriter, r *http.Request) {
  response.Result(w, true)
}

//POST map/character/
func (c *controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
  response.Result(w, true)
}

//PUT map/character/{uid}
func (c *controller) PutCharacter(w http.ResponseWriter, r *http.Request) {
  response.Result(w, true)
}

//DELETE map/character/{uid}
func (c *controller) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
  response.Result(w, true)
}
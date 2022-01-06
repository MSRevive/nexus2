package controller

import (
  "net/http"
  
  "github.com/msrevive/nexus2/response"
  
  //"github.com/gorilla/mux"
)

//GET map/character/
func (c *controller) GetAllCharacters(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}

//GET map/character/{steamid}
func (c *controller) GetCharacters(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}

//GET map/character/{steamid}/{slot}
func (c *controller) GetCharacter(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}

//GET map/character/id/{uid}
func (c *controller) GetCharacterByID(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}

//POST map/character/
func (c *controller) PostCharacter(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}

//PUT map/character/{uid}
func (c *controller) PutCharacter(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}

//DELETE map/character/{uid}
func (c *controller) DeleteCharacter(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}
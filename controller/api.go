package controller

import (
  "net/http"
  
  "github.com/msrevive/nexus2/response"
  
  //"github.com/gorilla/mux"
)

//GET map/{name}/{hash}
func (c *controller) GetMapVerify(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}

//GET ban/{steamid}
func (c *controller) GetBanVerify(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}

//GET sc/{hash}
func (c *controller) GetSCVerify(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}
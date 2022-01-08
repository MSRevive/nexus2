package controller

import (
  "net/http"
  
  "github.com/msrevive/nexus2/response"
  "github.com/msrevive/nexus2/service"
  "github.com/msrevive/nexus2/session"
  
  "github.com/gorilla/mux"
)

type controller struct {
  R *mux.Router
}

func New(router *mux.Router) *controller {
  return &controller{
    R: router,
  }
}

func (c *controller) TestRoot(w http.ResponseWriter, r *http.Request) {
  if session.Dbg {
    if err := service.New(r.Context()).Debug(); err != nil {
      response.BadRequest(w, err)
      return
    }
  }
  
  response.Result(w, true)
}
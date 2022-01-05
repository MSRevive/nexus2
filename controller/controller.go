package controller

import (
  "net/http"
  
  "github.com/msrevive/nexus2/response"
  
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

func (r *controller) TestRoot(res http.ResponseWriter, req *http.Request) {
  response.OK(res, "{}")
}
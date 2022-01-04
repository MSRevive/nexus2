package router

import (
  "net/http"
  
  "github.com/gorilla/mux"
)

func registerRoutes(r *mux.Router) {
  r.Subrouter()
}
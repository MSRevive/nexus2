package controller

import (
	"net/http"
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/msrevive/nexus2/internal/response"
	"github.com/msrevive/nexus2/internal/service"
	"github.com/msrevive/nexus2/internal/system"
)

type controller struct {
	R *mux.Router
	db *sql.DB
}

func New(router *mux.Router, db *sql.DB) *controller {
	return &controller{
		R: router,
		db: db,
	}
}

func (c *controller) TestRoot(w http.ResponseWriter, r *http.Request) {
	if system.HelperCfg.GetDebugMode() {
		if err := service.New(r.Context()).Debug(); err != nil {
			response.BadRequest(w, err)
			return
		}
	}

	response.Result(w, true)
}

type RouteHandler interface {
	ConfigureRoutes(router *mux.Router)
}

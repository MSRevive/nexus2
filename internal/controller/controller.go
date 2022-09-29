package controller

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/msrevive/nexus2/internal/response"
	"github.com/msrevive/nexus2/internal/service"
	"github.com/msrevive/nexus2/internal/system"
	"github.com/saintwish/auralog"
)

type controller struct {
	R   *mux.Router
	db  *sql.DB
	log *auralog.Logger
}

func New(router *mux.Router, db *sql.DB, log *auralog.Logger) *controller {
	return &controller{
		R:   router,
		db:  db,
		log: log,
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

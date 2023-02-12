package controller

import (
	"net/http"

	"github.com/gorilla/mux"
	//"github.com/msrevive/nexus2/pkg/response"
	//"github.com/msrevive/nexus2/internal/service"
	"github.com/msrevive/nexus2/cmd/app"
)

type controller struct {
	R *mux.Router
	App *app.App
}

func New(router *mux.Router, apps *app.App) *controller {
	return &controller{
		R: router,
		App: apps,
	}
}

func (c *controller) TestRoot(w http.ResponseWriter, r *http.Request) {
	// if system.Dbg {
	// 	if err := service.New(r.Context()).Debug(); err != nil {
	// 		response.BadRequest(w, err)
	// 		return
	// 	}
	// }

	// response.Result(w, true)
}

type RouteHandler interface {
	ConfigureRoutes(router *mux.Router)
}

package controller

import (
	"net/http"

	"github.com/msrevive/nexus2/cmd/app"
)

type controller struct {
	A *app.App
}

func New(a *app.App) *controller {
	return &controller{
		A: a,
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
package controller

import (
	"strconv"
	"net/http"

	"github.com/msrevive/nexus2/cmd/app"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Controller struct {
	a *app.App
}

func New(a *app.App) *Controller {
	return &Controller{
		a: a,
	}
}

//GET map/{name}/{hash}
func (c *Controller) GetMapVerify(w http.ResponseWriter, r *http.Request) {
	if !c.a.Config.Verify.EnforceMap {
		response.Result(w, true)
		return
	}
	
	name := chi.URLParam(r, "name")
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		c.a.Logger.API.Error("HTTP: failed to GetMapVerify", "map", name)
		response.BadRequest(w, err)
		return
	}
	
	if v,ok := c.a.List.Map.GetHas(name); ok && v == uint32(hash) {
		response.Result(w, true)
		return
	}
	
	c.a.Logger.API.Warn("Failed map verification", "IP", r.RemoteAddr, "map", name)
	response.Result(w, false)
	return
}

//GET ban/{steamid}
//in this case false means player isn't banned
func (c *Controller) GetBanVerify(w http.ResponseWriter, r *http.Request) {
	if !c.a.Config.Verify.EnforceBan {
		response.Result(w, false)
		return
	}
	
	steamid := chi.URLParam(r, "steamid")
	
	if ok := c.a.List.Ban.Has(steamid); ok {
		c.a.Logger.API.Warn("SteamID is banned from FN", "IP", r.RemoteAddr, "SteamID", steamid)
		response.Result(w, true)
		return
	}
	
	response.Result(w, false)
	return
}

//GET sc/{hash}
func (c *Controller) GetSCVerify(w http.ResponseWriter, r *http.Request) {
	if !c.a.Config.Verify.EnforceSC {
		response.Result(w, true)
		return
	}
	
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		c.a.Logger.API.Error("HTTP: failed to GetSCVerify")
		response.BadRequest(w, err)
		return
	}
	
	if c.a.Config.Verify.SCHash == uint32(hash) {
		response.Result(w, true)
		return
	}
	
	c.a.Logger.API.Warn("Failed SC verfication", "IP", r.RemoteAddr)
	response.Result(w, false)
}

//GET ping
func (c *Controller) GetPing(w http.ResponseWriter, r *http.Request) {
	response.Result(w, true)
}
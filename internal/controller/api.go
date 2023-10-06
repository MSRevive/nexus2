package controller

import (
	"strconv"
	"net/http"
	
	"github.com/msrevive/nexus2/pkg/response"
	
	"github.com/go-chi/chi/v5"
)

//GET map/{name}/{hash}
func (c *controller) GetMapVerify(w http.ResponseWriter, r *http.Request) {
	if !c.App.Config.Verify.EnforceMap {
		response.Result(w, true)
		return
	}
	
	name := chi.URLParam(r, "name")
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		response.BadRequest(w, err)
		return
	}
	
	if v,ok := c.App.List.Map.GetHas(name); ok && v == hash {
		c.App.Logger.API.Infof("%s passed map (%s) verfication.", r.RemoteAddr, name)
		response.Result(w, true)
		return
	}
	
	c.App.Logger.API.Warnf("%s failed map (%s) verfication.", r.RemoteAddr, name)
	response.Result(w, false)
	return
}

//GET ban/{steamid}
//in this case false means player isn't banned
func (c *controller) GetBanVerify(w http.ResponseWriter, r *http.Request) {
	if !c.App.Config.Verify.EnforceBan {
		response.Result(w, false)
		return
	}
	
	steamid := chi.URLParam(r, "steamid")
	
	if ok := c.App.List.Ban.Has(steamid); ok {
		c.App.Logger.API.Warnf("%s: player (%s) is banned!", r.RemoteAddr, steamid)
		response.Result(w, true)
		return
	}
	
	response.Result(w, false)
	return
}

//GET sc/{hash}
func (c *controller) GetSCVerify(w http.ResponseWriter, r *http.Request) {
	if !c.App.Config.Verify.EnforceSC {
		response.Result(w, true)
		return
	}
	
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		response.BadRequest(w, err)
		return
	}
	
	if c.App.Config.Verify.SCHash == hash {
		response.Result(w, true)
		return
	}
	
	c.App.Logger.API.Warnf("%s failed SC check!", r.RemoteAddr)
	response.Result(w, false)
}

//GET ping
func (c *controller) GetPing(w http.ResponseWriter, r *http.Request) {
	response.Result(w, true)
}
package controller

import (
	"strconv"
	"net/http"
	"log/slog"

	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/service"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/saintwish/kv/ccmap"
)

type Controller struct {
	logger *slog.Logger
	config *config.Config
	service *service.Service
	banList *ccmap.Cache[string, bool]
	mapList *ccmap.Cache[string, uint32]
	adminList *ccmap.Cache[string, bool]
}

func New(log *slog.Logger, cfg *config.Config, svr *service.Service, bans *ccmap.Cache[string, bool], maps *ccmap.Cache[string, uint32], admins *ccmap.Cache[string, bool]) *Controller {
	return &Controller{
		logger: log,
		config: cfg,
		service: svr,
		banList: bans,
		mapList: maps,
		adminList: admins,
	}
}

//GET map/{name}/{hash}
func (c *Controller) GetMapVerify(w http.ResponseWriter, r *http.Request) {
	if !c.config.Verify.EnforceMap {
		response.Result(w, true)
		return
	}
	
	name := chi.URLParam(r, "name")
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		c.logger.Error("HTTP: failed to GetMapVerify", "map", name)
		response.BadRequest(w, err)
		return
	}
	
	if v,ok := c.mapList.GetHas(name); ok && v == uint32(hash) {
		response.Result(w, true)
		return
	}
	
	c.logger.Warn("Failed map verification", "IP", r.RemoteAddr, "map", name)
	response.Result(w, false)
	return
}

//GET ban/{steamid}
//in this case false means player isn't banned
func (c *Controller) GetBanVerify(w http.ResponseWriter, r *http.Request) {
	if !c.config.Verify.EnforceBan {
		response.Result(w, false)
		return
	}
	
	steamid := chi.URLParam(r, "steamid")
	
	if ok := c.banList.Has(steamid); ok {
		c.logger.Warn("SteamID is banned from FN", "IP", r.RemoteAddr, "SteamID", steamid)
		response.Result(w, true)
		return
	}
	
	response.Result(w, false)
	return
}

//GET sc/{hash}
func (c *Controller) GetSCVerify(w http.ResponseWriter, r *http.Request) {
	if !c.config.Verify.EnforceSC {
		response.Result(w, true)
		return
	}
	
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		c.logger.Error("HTTP: failed to GetSCVerify")
		response.BadRequest(w, err)
		return
	}
	
	if c.config.Verify.SCHash == uint32(hash) {
		response.Result(w, true)
		return
	}
	
	c.logger.Warn("Failed SC verfication", "IP", r.RemoteAddr)
	response.Result(w, false)
}
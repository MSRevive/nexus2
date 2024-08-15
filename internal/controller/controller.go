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

type Options struct {
	BanList *ccmap.Cache[string, bool]
	MapList *ccmap.Cache[string, uint32]
	AdminList *ccmap.Cache[string, bool]
	ServerWinHash uint32
	ServerUnixHash uint32
	ScriptsHash uint32
}

type Controller struct {
	logger *slog.Logger
	config *config.Config
	service *service.Service
	banList *ccmap.Cache[string, bool]
	mapList *ccmap.Cache[string, uint32]
	adminList *ccmap.Cache[string, bool]

	serverWinHash uint32
	serverUnixHash uint32
	scriptsHash uint32
}

func New(service *service.Service, log *slog.Logger, cfg *config.Config, opts Options) *Controller {
	return &Controller{
		logger: log,
		service: service,
		config: cfg,
		banList: opts.BanList,
		mapList: opts.MapList,
		adminList: opts.AdminList,

		serverWinHash: opts.ServerWinHash,
		serverUnixHash: opts.ServerUnixHash,
		scriptsHash: opts.ScriptsHash,
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
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		c.logger.Error("HTTP: failed to GetSCVerify")
		response.BadRequest(w, err)
		return
	}
	
	if c.scriptsHash == uint32(hash) {
		response.Result(w, true)
		return
	}
	
	c.logger.Warn("Failed scripts verfication", "IP", r.RemoteAddr)
	response.Result(w, false)
}

//GET /server/{hash}
func (c *Controller) GetServerVerify(w http.ResponseWriter, r *http.Request) {
	response.Result(w, false)
}
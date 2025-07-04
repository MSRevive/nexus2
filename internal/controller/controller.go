package controller

import (
	"strconv"
	"net/http"
	"log/slog"

	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/service"
	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/response"

	"github.com/go-chi/chi/v5"
	"github.com/saintwish/kv/ccmap"
)

type Options struct {
	MapList *ccmap.Cache[string, uint32]
}

type Controller struct {
	logger *slog.Logger
	config *config.Config
	service *service.Service
	mapList *ccmap.Cache[string, uint32]
}

func New(service *service.Service, log *slog.Logger, cfg *config.Config, opts Options) *Controller {
	return &Controller{
		logger: log,
		service: service,
		config: cfg,
		mapList: opts.MapList,
	}
}

// DEPRECIATED
func (c *Controller) DepreciatedAPIVersion(w http.ResponseWriter, r *http.Request) {
	response.DepreciatedError(w)
}

func (c *Controller) NotAvailable(w http.ResponseWriter, r *http.Request) {
	response.NotAvailable(w)
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

//GET sc/{hash}
func (c *Controller) GetSCVerify(w http.ResponseWriter, r *http.Request) {
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		c.logger.Error("HTTP: failed to GetSCVerify")
		response.BadRequest(w, err)
		return
	}

	if !c.config.Verify.EnforceBins {
		response.Result(w, true)
		return
	}
	
	if (c.config.Verify.ScriptsHash == 0) || (uint32(hash) == c.config.Verify.ScriptsHash) {
		response.Result(w, true)
		return
	}
	
	c.logger.Warn("Failed scripts verfication", "IP", r.RemoteAddr)
	response.Result(w, false)
}

//GET /server/{hash}
func (c *Controller) GetServerVerify(w http.ResponseWriter, r *http.Request) {
	hash, err := strconv.ParseUint(chi.URLParam(r, "hash"), 10, 32)
	if err != nil {
		c.logger.Error("HTTP: failed to GetServerVerify")
		response.BadRequest(w, err)
		return
	}

	if !c.config.Verify.EnforceBins {
		response.Result(w, true)
		return
	}

	if ((c.config.Verify.ServerUnixHash == 0) || (c.config.Verify.ServerWinHash == 0)) || ((uint32(hash) == c.config.Verify.ServerUnixHash) || (uint32(hash) == c.config.Verify.ServerWinHash)) {
		response.Result(w, true)
		return
	}

	c.logger.Warn("Failed server verfication", "IP", r.RemoteAddr)
	response.Result(w, false)
}

//GET ban/{steamid}
//in this case false means player isn't banned
func (c *Controller) GetBanVerify(w http.ResponseWriter, r *http.Request) {
	if !c.config.Verify.EnforceBan {
		response.Result(w, false)
		return
	}
	
	steamid := chi.URLParam(r, "steamid")
	
	flags, err := c.service.GetUserFlags(steamid)
	if err != nil {
		c.logger.Error("Unable to get user flags from SteamID", "IP", r.RemoteAddr, "SteamID", steamid)
		response.GenericError(w)
		return
	}

	if flags.HasFlag(bitmask.BANNED) {
		c.logger.Warn("SteamID is banned from FN", "IP", r.RemoteAddr, "SteamID", steamid)
		response.Result(w, true)
		return
	}
	
	response.Result(w, false)
	return
}
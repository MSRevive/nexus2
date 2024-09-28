package controller

import (
	"strconv"
	"net/http"
	"log/slog"

	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/service"
	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/saintwish/kv/ccmap"
)

//POST user/ban/{steamid}
//in this case false means player isn't banned
func (c *Controller) PostBanSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	if err := service.AddUserFlag(steamid, bitmask.BANNED); err != nil {
		response.Error(w, err)
		return
	}
	
	response.Result(w, false)
	return
}
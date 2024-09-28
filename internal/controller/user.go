package controller

import (
	"net/http"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
)

//POST user/ban/{steamid}
//in this case false means player isn't banned
func (c *Controller) PostBanSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	if err := c.service.AddUserFlag(steamid, bitmask.BANNED); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}
	
	response.Result(w, false)
	return
}
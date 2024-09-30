package controller

import (
	"net/http"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/response"

	"github.com/go-chi/chi/v5"
)

func (c *Controller) GetUser(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")

	user, err := c.service.GetUser(steamid)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OK(w, user)
	return
}

func (c *Controller) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.service.GetAllUsers()
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OK(w, users)
	return
}

//PATCH user/ban/{steamid}
func (c *Controller) PatchBanSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	if err := c.service.AddUserFlag(steamid, bitmask.BANNED); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}
	
	response.OKNoContent(w)
	return
}

//PATCH user/unban/{steamid}
func (c *Controller) PatchUnBanSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	if err := c.service.RemoveUserFlag(steamid, bitmask.BANNED); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}
	
	response.OKNoContent(w)
	return
}

//PATCH user/admin/{steamid}
func (c *Controller) PatchAdminSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	if err := c.service.AddUserFlag(steamid, bitmask.ADMIN); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}
	
	response.OKNoContent(w)
	return
}

//PATCH user/unadmin/{steamid}
func (c *Controller) PatchUnAdminSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	if err := c.service.RemoveUserFlag(steamid, bitmask.ADMIN); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}
	
	response.OKNoContent(w)
	return
}

//PATCH user/donor/{steamid}
func (c *Controller) PatchDonorSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	if err := c.service.AddUserFlag(steamid, bitmask.DONOR); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}
	
	response.OKNoContent(w)
	return
}

//PATCH user/undonor/{steamid}
func (c *Controller) PatchUnDonorSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	if err := c.service.RemoveUserFlag(steamid, bitmask.DONOR); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}
	
	response.OKNoContent(w)
	return
}

//PATCH user/isdonor/{steamid}
func (c *Controller) GetIsDonorSteamID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	
	flags, err := c.service.GetUserFlags(steamid)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}
	
	if flags.HasFlag(bitmask.DONOR) {
		response.OK(w, true)
	}
	response.OK(w, false)
	return
}

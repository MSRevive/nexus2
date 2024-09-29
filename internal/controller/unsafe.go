package controller

import (
	"net/http"
	"strconv"

	"github.com/msrevive/nexus2/internal/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PATCH /unsafe/character/move/{uuid}/to/{steamid:[0-9]+}/{slot:[0-9]}
func (c *Controller) UnsafeMoveCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}
	steamid := chi.URLParam(r, "steamid")
	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	newUID, err := c.service.MoveCharacter(uid, steamid, slot)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.BadRequest(w, err)
		return
	}

	response.OK(w, newUID.String())
}

// PATCH /unsafe/character/copy/{uuid}/to/{steamid:[0-9]+}/{slot:[0-9]}
func (c *Controller) UnsafeCopyCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}
	steamid := chi.URLParam(r, "steamid")
	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	newUID, err := c.service.CopyCharacter(uid, steamid, slot)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.BadRequest(w, err)
		return
	}

	response.OK(w, newUID.String())
}

// DELETE /unsafe/character/delete/{uuid}
func (c *Controller) UnsafeDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	if err := c.service.HardDeleteCharacter(uid); err != nil {
		c.logger.Error("service failed", "error", err)
		response.BadRequest(w, err)
		return
	}

	response.OK(w, uid)
}
package controller

import (
	"net/http"
	"strconv"

	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GET /rollback/character/{uuid}
func (c *Controller) GetCharacterVersions(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	data, err := c.service.GetCharacterVersions(uid)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OK(w, data)
}

// PATCH /rollback/character/{uuid}/latest
func (c *Controller) RollbackCharToLatest(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	// 0 will be the first backup
	if err := c.service.RollbackCharacterToLatest(uid); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OKNoContent(w)
}

// PATCH /rollback/character/{uuid}/{version:[0-9]+}
func (c *Controller) RollbackCharToVersion(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}
	ver, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	if err := c.service.RollbackCharacter(uid, ver); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OKNoContent(w)
}

// DELETE /rollback/character/{uuid}
func (c *Controller) DeleteCharRollbacks(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	if err := c.service.DeleteCharacterVersions(uid); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OKNoContent(w)
}
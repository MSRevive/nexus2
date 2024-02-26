package controller

import (
	"fmt"
	"net/http"
	"io"
	"strconv"

	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GET /character/lookup/{steamid:[0-9]+}/{slot:[0-9]}
func (c *Controller) LookUpCharacterID(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	uid, err := c.service.LookUpCharacterID(steamid, slot)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OK(w, uid.String())
}

// PATCH /character/restore/{uuid}
func (c *Controller) RestoreCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	if err := c.service.RestoreCharacter(uid); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OK(w, uid.String())
}

// GET /character/export/{uuid}
func (c *Controller) ExportCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	char, err := c.service.GetCharacterByID(uid)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	file, path, err := payload.GenerateCharFile(char.SteamID, char.Slot, char.Data.Data)
	if err != nil {
		c.logger.Error("character file generation failed", "error", err)
		response.Error(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", path))
	io.Copy(w, file)
}

// GET /character/{uuid}
func (c *Controller) GetCharacterByIDExternal(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}
	isBanned := false;
	isAdmin := false;

	char, err := c.service.GetCharacterByID(uid)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OKChar(w, isBanned, isAdmin, char)
}
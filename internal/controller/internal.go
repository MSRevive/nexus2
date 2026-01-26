package controller

import (
	"net/http"
	"io"
	"bytes"
	"strconv"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/internal/response"
	"github.com/msrevive/nexus2/pkg/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// POST /internal/character/
func (c *Controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
	// copy the content of body into a new buffer for processing
	var buf bytes.Buffer
	if size, err := io.Copy(&buf, r.Body); err != nil {
		c.logger.Error("failed to copy body", "error", err, "size", size, "expectedSize", r.ContentLength)
		response.Error(w, err)
		return
	}
	body := buf.Bytes()

	var char payload.Character
	if err := utils.ProcessJSON(body, &char); err != nil {
		c.logger.Error("failed to parse JSON", "body", body, "error", err)
		response.BadRequest(w, err)
		return
	}

	uid, flags, err := c.service.NewCharacter(char); 
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.Created(w, payload.CharacterCreate{
		ID: uid,
		Flags: flags,
	})
}

// PUT /internal/character/{uuid}
func (c *Controller) PutCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	// copy the content of body into a new buffer for processing
	var buf bytes.Buffer
	if size, err := io.Copy(&buf, r.Body); err != nil {
		c.logger.Error("failed to copy body", "error", err, "size", size, "expectedSize", r.ContentLength)
		response.Error(w, err)
		return
	}
	body := buf.Bytes()

	var char payload.Character

	if err := utils.ProcessJSON(body, &char); err != nil {
		c.logger.Error("failed to parse JSON", "uuid", uid, "body", body, "error", err)
		response.BadRequest(w, err)
		return
	}

	if err := c.service.UpdateCharacter(uid, char); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OK(w, uid.String())
}

// GET /internal/character/{steamid:[0-9]+}/{slot:[0-9]}
func (c *Controller) GetCharacter(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	char, flags, err := c.service.GetCharacter(steamid, slot)
	if err == database.ErrNoDocument {
		c.logger.Warn("service warning", "error", err)
		response.OKNoContent(w)
		return
	}else if err != nil {
		c.logger.Error("service error", "error", err)
		response.Error(w, err)
		return
	}

	response.OKChar(w, payload.Character{
		ID: char.ID,
		SteamID: char.SteamID,
		Slot: char.Slot,
		Size: char.Data.Size,
		Data: char.Data.Data,
		Flags: flags,
	})
}

// DELETE /internal/character/{uuid}
func (c *Controller) SoftDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	if err := c.service.SoftDeleteCharacter(uid, c.config.Char.DeletedExpireTime); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OK(w, uid.String())
}
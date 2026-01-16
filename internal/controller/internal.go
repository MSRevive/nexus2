package controller

import (
	"fmt"
	"net/http"
	"io"
	"bytes"
	"strconv"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/internal/response"

	"github.com/go-chi/chi/v5"
	json "github.com/sugawarayuuta/sonnet"
	"github.com/google/uuid"
)

// POST /internal/character/
func (c *Controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
	var char payload.Character
	if err := json.NewDecoder(r.Body).Decode(&char); err != nil {
		c.logger.Debug("POST character body sent", "data", r.Body)
		var buf bytes.Buffer

		if size, err := io.Copy(&buf, r.Body); err != nil {
			c.logger.Error("failed to copy body", "error", err, "size", size, "expectedSize", r.ContentLength)
			response.Error(w, err)
			return
		}

		data := buf.Bytes()
		c.logger.Debug("character data sent", "data", data)
		var errln error
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			errln = fmt.Errorf("%w ~ error near '%s' (offset %d)", err, string(data[jsonErr.Offset-1:]), jsonErr.Offset)
		}

		c.logger.Error("failed to parse data", "error", errln)
		if errln == nil {
			response.GenericError(w)
			return
		}

		response.BadRequest(w, errln)
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

	var char payload.Character
	if err := json.NewDecoder(r.Body).Decode(&char); err != nil {
		c.logger.Debug("PUT character body sent", "data", r.Body)
		var buf bytes.Buffer

		if size, err := io.Copy(&buf, r.Body); err != nil {
			c.logger.Error("failed to copy body", "error", err, "size", size, "expectedSize", r.ContentLength)
			response.Error(w, err)
			return
		}

		data := buf.Bytes()
		c.logger.Debug("character data sent", "data", data)
		var errln error
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			errln = fmt.Errorf("%w ~ error near '%s' (offset %d)", err, data, jsonErr.Offset)
		}

		c.logger.Error("failed to parse data", "error", errln)
		if errln == nil {
			response.GenericError(w)
			return
		}

		response.BadRequest(w, errln)
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
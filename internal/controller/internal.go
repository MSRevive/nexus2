package controller

import (
	"fmt"
	"net/http"
	"io"
	"bytes"
	"strconv"

	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/internal/response"

	"github.com/go-chi/chi/v5"
	json "github.com/goccy/go-json"
	"github.com/google/uuid"
)

// POST /internal/character/
func (c *Controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
	var char payload.Character
	if err := json.NewDecoder(r.Body).Decode(&char); err != nil {
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
			problemPart := data[jsonErr.Offset-10 : jsonErr.Offset+10]
			errln = fmt.Errorf("%w ~ error near '%s' (offset %d)", err, problemPart, jsonErr.Offset)
		}

		c.logger.Error("failed to parse data", "error", errln)
		if errln == nil {
			response.GenericError(w)
			return
		}

		response.BadRequest(w, errln)
		return
	}

	uid, err := c.service.NewCharacter(char); 
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.Created(w, uid.String())
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
			problemPart := data[jsonErr.Offset-10 : jsonErr.Offset+10]
			errln = fmt.Errorf("%w ~ error near '%s' (offset %d)", err, problemPart, jsonErr.Offset)
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
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OKChar(w, payload.Character{
		ID: char.ID,
		SteamID: char.SteamID,
		Slot: char.Slot,
		Size: char.Data.Size,
		Data: char.Data.Data,
	}, flags)
}

// GET /internal/character/{steamid:[0-9]+}
func (c *Controller) GetCharacters(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")

	chars, flags, err := c.service.GetCharacters(steamid)
	if err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OKChar(w, chars, flags)
}

// DELETE /internal/character/{uuid}
func (c *Controller) SoftDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		c.logger.Error("controller: bad request", "error", err)
		response.BadRequest(w, err)
		return
	}

	if err := c.service.SoftDeleteCharacter(uid, config.Char.DeletedExpireTime); err != nil {
		c.logger.Error("service failed", "error", err)
		response.Error(w, err)
		return
	}

	response.OK(w, uid.String())
}
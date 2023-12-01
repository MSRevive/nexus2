package controller

import (
	"fmt"
	"net/http"
	"io"
	"bytes"

	"github.com/msrevive/nexus2/internal/database/schema"
	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
	json "github.com/goccy/go-json"
	"github.com/google/uuid"
)

//POST /character/
func (c *Controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
	var newChar schema.CharacterData
	if err := json.NewDecoder(r.Body).Decode(&newChar); err != nil {
		var buf bytes.Buffer

		if size, err := io.Copy(&buf, r.Body); err != nil {
			c.a.Logger.API.Error("failed to copy body", "error", err, "size", size, "expectedSize", r.ContentLength)
			response.BadRequest(w, err)
			return
		}

		data := buf.Bytes()
		c.a.Logger.API.Debug("character data sent", "data", data)
		var errln error
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			problemPart := data[jsonErr.Offset-10 : jsonErr.Offset+10]
			errln = fmt.Errorf("%w ~ error near '%s' (offset %d)", err, problemPart, jsonErr.Offset)
		}

		c.a.Logger.API.Error("failed to parse data", "error", errln)
		if errln == nil {
			response.InternalServerError(w)
			return
		}

		response.BadRequest(w, errln)
		return
	}

	response.OK(w, newChar)
}

//PUT /character/{uid}
func (c *Controller) PutCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.a.Logger.API.Error("error parsing UUID", "error", err)
		response.BadRequest(w, err)
		return
	}

	var updatedChar schema.CharacterData
	if err := json.NewDecoder(r.Body).Decode(&updatedChar); err != nil {
		var buf bytes.Buffer

		if size, err := io.Copy(&buf, r.Body); err != nil {
			c.a.Logger.API.Error("failed to copy body", "error", err, "size", size, "expectedSize", r.ContentLength)
			response.BadRequest(w, err)
			return
		}

		data := buf.Bytes()
		c.a.Logger.API.Debug("character data sent", "data", data)
		var errln error
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			problemPart := data[jsonErr.Offset-10 : jsonErr.Offset+10]
			errln = fmt.Errorf("%w ~ error near '%s' (offset %d)", err, problemPart, jsonErr.Offset)
		}

		c.a.Logger.API.Error("failed to parse data", "error", errln)
		if errln == nil {
			response.InternalServerError(w)
			return
		}

		response.BadRequest(w, errln)
		return
	}

	response.OK(w, uid)
}

func (c *Controller) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.a.Logger.API.Error("error parsing UUID", "error", err)
		response.BadRequest(w, err)
		return
	}

	response.OK(w, uid)
}
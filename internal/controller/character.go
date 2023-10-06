package controller

import (
	"fmt"
	"io"
	"bytes"
	"net/http"
	"strconv"

	"github.com/msrevive/nexus2/ent"
	"github.com/msrevive/nexus2/pkg/helper"
	"github.com/msrevive/nexus2/pkg/response"
	"github.com/msrevive/nexus2/internal/service"

	json "github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
)

//GET /character/
func (c *controller) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
	chars, err := service.New(r.Context(), c.App).CharactersGetAll()
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	response.OK(w, chars)
}

//GET /character/{steamid}
func (c *controller) GetCharacters(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	var isBanned bool = false
	var isAdmin bool = false

	chars, err := service.New(r.Context(), c.App).CharactersGetBySteamid(steamid)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if ok := c.App.List.Ban.Has(steamid); ok && c.App.Config.Verify.EnforceBan {
		isBanned = true
	}

	if ok := c.App.List.Admin.Has(steamid); ok {
		isAdmin = true
	}

	response.OKChar(w, isBanned, isAdmin, chars)
}

//GET /character/{steamid}/{slot}
func (c *controller) GetCharacter(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}
	var isBanned bool = false
	var isAdmin bool = false

	char, err := service.New(r.Context(), c.App).CharacterGetBySteamidSlot(steamid, slot)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if ok := c.App.List.Ban.Has(steamid); ok && c.App.Config.Verify.EnforceBan {
		isBanned = true
	}

	if ok := c.App.List.Admin.Has(steamid); ok {
		isAdmin = true
	}

	response.OKChar(w, isBanned, isAdmin, char)
}

//GET /character/export/{steamid}/{slot}
func (c *controller) ExportCharacter(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterGetBySteamidSlot(steamid, slot)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	file, path, err := helper.GenerateCharFile(steamid, slot, char.Data)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", path))
	io.Copy(w, file)
}

//GET /character/id/{uid}
func (c *controller) GetCharacterByID(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}
	var isBanned bool = false
	var isAdmin bool = false

	char, err := service.New(r.Context(), c.App).CharacterGetByID(uid)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if ok := c.App.List.Ban.Has(char.SteamID); ok && c.App.Config.Verify.EnforceBan {
		isBanned = true
	}

	if ok := c.App.List.Admin.Has(char.SteamID); ok {
		isAdmin = true
	}

	response.OKChar(w, isBanned, isAdmin, char)
}

//POST /character/
func (c *controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newChar ent.DeprecatedCharacter
	if err := json.NewDecoder(r.Body).Decode(&newChar); err != nil {
		var buf bytes.Buffer
		if size, err := io.Copy(&buf, r.Body); err != nil {
			c.App.Logger.API.Errorf("failed to copy body: %s copied (%d) expected (%d)", err, size, r.ContentLength)
			response.BadRequest(w, err)
			return
		}

		data := buf.Bytes()
		
		c.App.Logger.API.Debugln(data)
		var errln error
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			problemPart := data[jsonErr.Offset-10 : jsonErr.Offset+10]
			errln = fmt.Errorf("%w ~ error near '%s' (offset %d)", err, problemPart, jsonErr.Offset)
		}

		c.App.Logger.API.Errorln(errln)
		if errln == nil {
			response.InternalServerError(w)
			return
		}
		response.BadRequest(w, errln)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterCreate(newChar)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//PUT /character/{uid}
func (c *controller) PutCharacter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	var updateChar ent.DeprecatedCharacter
	if err := json.NewDecoder(r.Body).Decode(&updateChar); err != nil {
		var buf bytes.Buffer
		if size, err := io.Copy(&buf, r.Body); err != nil {
			c.App.Logger.API.Errorf("failed to copy body: %s copied (%d) expected (%d)", err, size, r.ContentLength)
			response.BadRequest(w, err)
			return
		}

		data := buf.Bytes()

		c.App.Logger.API.Debugln(data)
		var errln error
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			problemPart := data[jsonErr.Offset-10 : jsonErr.Offset+10]
			errln = fmt.Errorf("%w ~ error near '%s' (offset %d)", err, problemPart, jsonErr.Offset)
		}

		c.App.Logger.API.Errorln(errln)
		if errln == nil {
			response.InternalServerError(w)
			return
		}
		response.BadRequest(w, errln)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterUpdate(uid, updateChar)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//DELETE /character/{uid}
func (c *controller) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	err = service.New(r.Context(), c.App).CharacterDelete(uid)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, uid)
}

//PATCH /character/{uid}/restore
func (c *controller) RestoreCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterRestore(uid)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//PATCH /character/{steamid}/{slot}/restore
func (c *controller) RestoreCharacterBySteamID(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterRestoreBySteamID(sid, slot)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//GET /character/{steamid}/{slot}/versions
func (c *controller) CharacterVersions(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterVersions(sid, slot)
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//PATCH /character/transfer/{uid}/to/{steamid}/{slot}
func (c *controller) CharacterTransfer(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if err := service.New(r.Context(), c.App).CharacterTransfer(uid, sid, slot); err != nil {
		c.App.Logger.API.Errorln(err)
		response.Error(w, err)
		return
	}
}

//PATCH /character/copy/{uid}/to/{steamid}/{slot}
func (c *controller) CharacterCopy(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.Logger.API.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if err := service.New(r.Context(), c.App).CharacterCopy(uid, sid, slot); err != nil {
		c.App.Logger.API.Errorln(err)
		response.Error(w, err)
		return
	}
}
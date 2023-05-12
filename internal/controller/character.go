package controller

import (
	"fmt"
	"io"
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
		c.App.LogAPI.Errorln(err)
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
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if _, ok := c.App.BanList[steamid]; ok && c.App.Config.Verify.EnforceBan {
		isBanned = true
	}

	if _, ok := c.App.AdminList[steamid]; ok {
		isAdmin = true
	}

	response.OKChar(w, isBanned, isAdmin, chars)
}

//GET /character/{steamid}/{slot}
func (c *controller) GetCharacter(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}
	var isBanned bool = false
	var isAdmin bool = false

	char, err := service.New(r.Context(), c.App).CharacterGetBySteamidSlot(steamid, slot)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if _, ok := c.App.BanList[steamid]; ok && c.App.Config.Verify.EnforceBan {
		isBanned = true
	}

	if _, ok := c.App.AdminList[steamid]; ok {
		isAdmin = true
	}

	response.OKChar(w, isBanned, isAdmin, char)
}

//GET /character/export/{steamid}/{slot}
func (c *controller) ExportCharacter(w http.ResponseWriter, r *http.Request) {
	steamid := chi.URLParam(r, "steamid")
	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterGetBySteamidSlot(steamid, slot)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	file, path, err := helper.GenerateCharFile(steamid, slot, char.Data)
	if err != nil {
		c.App.LogAPI.Errorln(err)
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
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}
	var isBanned bool = false
	var isAdmin bool = false

	char, err := service.New(r.Context(), c.App).CharacterGetByID(uid)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if _, ok := c.App.BanList[char.Steamid]; ok && c.App.Config.Verify.EnforceBan {
		isBanned = true
	}

	if _, ok := c.App.AdminList[char.Steamid]; ok {
		isAdmin = true
	}

	response.OKChar(w, isBanned, isAdmin, char)
}

//POST /character/
func (c *controller) PostCharacter(w http.ResponseWriter, r *http.Request) {
	var newChar ent.DeprecatedCharacter
	err := json.NewDecoder(r.Body).Decode(&newChar)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterCreate(newChar)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//PUT /character/{uid}
func (c *controller) PutCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	var updateChar ent.DeprecatedCharacter
	err = json.NewDecoder(r.Body).Decode(&updateChar)
	if err != nil {
		c.App.LogAPI.Warnln(r.Body)
		c.App.LogAPI.Errorln(err)
		c.App.LogAPI.Traceln()
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterUpdate(uid, updateChar)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//DELETE /character/{uid}
func (c *controller) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	err = service.New(r.Context(), c.App).CharacterDelete(uid)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, uid)
}

//PATCH /character/{uid}/restore
func (c *controller) RestoreCharacter(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterRestore(uid)
	if err != nil {
		c.App.LogAPI.Errorln(err)
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
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterRestoreBySteamID(sid, slot)
	if err != nil {
		c.App.LogAPI.Errorln(err)
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
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterVersions(sid, slot)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//PATCH /character/transfer/{uid}/to/{steamid}/{slot}
func (c *controller) CharacterTransfer(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if err := service.New(r.Context(), c.App).CharacterTransfer(uid, sid, slot); err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}
}

//PATCH /character/copy/{uid}/to/{steamid}/{slot}
func (c *controller) CharacterCopy(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	if err := service.New(r.Context(), c.App).CharacterCopy(uid, sid, slot); err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}
}
package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/msrevive/nexus2/ent"
	"github.com/msrevive/nexus2/pkg/helper"
	"github.com/msrevive/nexus2/pkg/response"
	"github.com/msrevive/nexus2/internal/service"
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
	vars := mux.Vars(r)
	steamid := vars["steamid"]
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
	vars := mux.Vars(r)
	steamid := vars["steamid"]
	slot, err := strconv.Atoi(vars["slot"])
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
	vars := mux.Vars(r)
	steamid := vars["steamid"]
	slot, err := strconv.Atoi(vars["slot"])
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
	vars := mux.Vars(r)
	uid, err := uuid.Parse(vars["uid"])
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
	vars := mux.Vars(r)
	uid, err := uuid.Parse(vars["uid"])
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	var updateChar ent.DeprecatedCharacter
	err = json.NewDecoder(r.Body).Decode(&updateChar)
	if err != nil {
		c.App.LogAPI.Errorln(err)
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
	vars := mux.Vars(r)
	uid, err := uuid.Parse(vars["uid"])
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
	vars := mux.Vars(r)
	uid, err := uuid.Parse(vars["uid"])
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

//GET /character/{steamid}/{slot}/versions
func (c *controller) CharacterVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, ok := vars["steamid"]
	if !ok {
		err := errors.New("steamid not found")
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	slot, err := strconv.Atoi(vars["slot"])
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

//PATCH /character/{steamid}/{slot}/rollback/{version}
func (c *controller) RollbackCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, ok := vars["steamid"]
	if !ok {
		err := errors.New("steamid not found")
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	slot, err := strconv.Atoi(vars["slot"])
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	version, err := strconv.Atoi(vars["version"])
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterRollback(sid, slot, version)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//PATCH /character/{steamid}/{slot}/rollback/latest
func (c *controller) RollbackLatestCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, ok := vars["steamid"]
	if !ok {
		err := errors.New("steamid not found")
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	slot, err := strconv.Atoi(vars["slot"])
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	char, err := service.New(r.Context(), c.App).CharacterRollbackLatest(sid, slot)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, char)
}

//DELETE /character/{steamid}/{slot}
func (c *controller) DeleteRollbacksCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sid, ok := vars["steamid"]
	if !ok {
		err := errors.New("steamid not found")
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	slot, err := strconv.Atoi(vars["slot"])
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	err = service.New(r.Context(), c.App).CharacterDeleteRollbacks(sid, slot)
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.Error(w, err)
		return
	}

	response.OK(w, sid)
}
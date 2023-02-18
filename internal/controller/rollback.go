package controller

import (
	"net/http"
	"strconv"

	"github.com/msrevive/nexus2/pkg/response"
	"github.com/msrevive/nexus2/internal/service"

	"github.com/go-chi/chi/v5"
)

//PATCH /character/rollback/{steamid}/{slot}/{version}
func (c *controller) RollbackCharacter(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
	if err != nil {
		c.App.LogAPI.Errorln(err)
		response.BadRequest(w, err)
		return
	}

	version, err := strconv.Atoi(chi.URLParam(r, "version"))
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

//PATCH /character/rollback/{steamid}/{slot}/latest
func (c *controller) RollbackLatestCharacter(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
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

//DELETE /character/rollback/{steamid}/{slot}
func (c *controller) DeleteRollbacksCharacter(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "steamid")

	slot, err := strconv.Atoi(chi.URLParam(r, "slot"))
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
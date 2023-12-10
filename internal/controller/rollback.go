package controller

import (
	"net/http"

	"github.com/msrevive/nexus2/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PATCH /character/restore/{uuid}
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
		response.BadRequest(w, err)
		return
	}

	response.OK(w, data)
}
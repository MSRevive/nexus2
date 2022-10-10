package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/controller"
	"github.com/msrevive/nexus2/internal/middleware"
	"github.com/saintwish/auralog"
)

func NewRouter(cfg *config.ApiConfig, log *auralog.Logger, db *sqlx.DB) {
	router := mux.NewRouter()

	//middleware
	mw := middleware.New(cfg, log)
	router.Use(mw.PanicRecovery)
	router.Use(mw.LogRequests)
	if cfg.RateLimit.Enable {
		router.Use(mw.RateLimit)
	}

	// API Routes
	ac := controller.NewApiController(router.PathPrefix(cfg.Core.RootPath).Subrouter(), db, log)
	api.R.HandleFunc("/ping", mw.Lv2Auth(api.GetPing)).Methods(http.MethodGet)
	api.R.HandleFunc("/map/{name}/{hash}", mw.Lv1Auth(api.GetMapVerify)).Methods(http.MethodGet)
	api.R.HandleFunc("/ban/{steamid:[0-9]+}", mw.Lv1Auth(api.GetBanVerify)).Methods(http.MethodGet)
	api.R.HandleFunc("/sc/{hash}", mw.Lv1Auth(api.GetSCVerify)).Methods(http.MethodGet)

	// Character Routes
	cc := controller.NewCharacterController(router.PathPrefix(cfg.Core.RootPath+"/character").Subrouter(), db, log)
	capi.R.HandleFunc("/", mw.Lv1Auth(capi.GetAllCharacters)).Methods(http.MethodGet)
	capi.R.HandleFunc("/id/{uid}", mw.Lv1Auth(capi.GetCharacterByID)).Methods(http.MethodGet)
	capi.R.HandleFunc("/{steamid:[0-9]+}", mw.Lv1Auth(capi.GetCharacters)).Methods(http.MethodGet)
	capi.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(capi.GetCharacter)).Methods(http.MethodGet)
	capi.R.HandleFunc("/export/{steamid:[0-9]+}/{slot:[0-9]}", mw.Lv1Auth(capi.ExportCharacter)).Methods(http.MethodGet)
	capi.R.HandleFunc("/", mw.Lv2Auth(capi.PostCharacter)).Methods(http.MethodPost)
	capi.R.HandleFunc("/{uid}", mw.Lv2Auth(capi.PutCharacter)).Methods(http.MethodPut)
	capi.R.HandleFunc("/{uid}", mw.Lv2Auth(capi.DeleteCharacter)).Methods(http.MethodDelete)
	capi.R.HandleFunc("/{uid}/restore", mw.Lv1Auth(capi.RestoreCharacter)).Methods(http.MethodPatch)
	capi.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/versions", mw.Lv1Auth(capi.CharacterVersions)).Methods(http.MethodGet)
	capi.R.HandleFunc("/{steamid:[0-9]+}/{slot:[0-9]}/rollback/{version:[0-9]+}", mw.Lv1Auth(capi.RollbackCharacter)).Methods(http.MethodPatch)

}

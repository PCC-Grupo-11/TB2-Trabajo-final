package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/auth"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/clients"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/vectorizer"
)

type Handler struct {
	Repo    *storage.Repository
	Cache   *storage.Cache
	Vec     *vectorizer.Vectorizer
	LB      *clients.LoadBalancer
	AuthSvc *auth.Auth
	Cfg     *config.APIConfig
}

func New(
	repo *storage.Repository,
	cache *storage.Cache,
	vec *vectorizer.Vectorizer,
	lb *clients.LoadBalancer,
	authSvc *auth.Auth,
	cfg *config.APIConfig,
) *Handler {
	return &Handler{
		Repo:    repo,
		Cache:   cache,
		Vec:     vec,
		LB:      lb,
		AuthSvc: authSvc,
		Cfg:     cfg,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	mongoStatus := "OK"
	if h.Repo != nil {
		if err := h.Repo.Ping(ctx); err != nil {
			mongoStatus = "FAIL"
		}
	} else {
		mongoStatus = "FAIL"
	}

	redisStatus := "OK"
	if h.Cache != nil {
		if err := h.Cache.Ping(ctx); err != nil {
			redisStatus = "FAIL"
		}
	} else {
		redisStatus = "FAIL"
	}

	status := http.StatusOK
	if mongoStatus == "FAIL" || redisStatus == "FAIL" {
		status = http.StatusServiceUnavailable
	}

	protocol.WriteJSON(w, status, map[string]string{
		"endpoint": "OK",
		"MongoDB":  mongoStatus,
		"Redis":    redisStatus,
	})
}

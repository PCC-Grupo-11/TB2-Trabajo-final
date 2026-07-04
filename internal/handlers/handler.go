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

func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
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

func (h *Handler) ResetMetrics(w http.ResponseWriter, r *http.Request) {
	if h.Cache == nil {
		protocol.WriteJSON(w, http.StatusServiceUnavailable, protocol.ErrorResponse{Error: "redis unavailable"})
		return
	}
	if err := h.Cache.ResetMetrics(r.Context()); err != nil {
		protocol.WriteJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "failed to reset metrics"})
		return
	}
	protocol.WriteJSON(w, http.StatusOK, map[string]string{"status": "metrics reset"})
}

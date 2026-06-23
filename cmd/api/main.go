package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/auth"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/clients"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/handlers"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/vectorizer"
)

func main() {
	cfg, err := config.LoadAPIConfig()
	if err != nil {
		logger.Error("invalid api config", "error", err)
		os.Exit(1)
	}

	logger.Info("api server starting",
		"port", cfg.Port,
		"mongo_uri", cfg.MongoURI,
		"redis_addr", cfg.RedisAddr,
	)

	ctx := context.Background()

	repo, err := storage.NewRepository(ctx, cfg.MongoURI)
	if err != nil {
		logger.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer repo.Close()
	logger.Info("connected to MongoDB")

	cache, err := storage.NewCache(cfg.RedisAddr, cfg.RedisTTL)
	if err != nil {
		logger.Warn("redis unavailable, continuing without cache", "error", err)
	}
	if cache != nil {
		defer cache.Close()
		logger.Info("connected to Redis", "addr", cfg.RedisAddr)
	}

	vec, err := vectorizer.New(cfg.MappingsDir)
	if err != nil {
		logger.Error("failed to load vectorizer", "error", err)
		os.Exit(1)
	}
	logger.Info("vectorizer loaded", "mappings_dir", cfg.MappingsDir)

	lb := clients.NewLoadBalancer(cfg.InferenceAddrs, cfg.InferenceTimeout)
	authSvc := auth.New(repo, cfg.JWTSecret, cfg.JWTExpiration)
	h := handlers.New(repo, cache, vec, lb, authSvc, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.Handle("POST /predict", authSvc.Middleware(http.HandlerFunc(h.Predict)))
	mux.Handle("POST /predict/bulk", authSvc.Middleware(http.HandlerFunc(h.PredictBulk)))
	mux.Handle("GET /metrics", authSvc.Middleware(http.HandlerFunc(h.Metrics)))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		logger.Info("http server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server")
	server.Shutdown(context.Background())
}

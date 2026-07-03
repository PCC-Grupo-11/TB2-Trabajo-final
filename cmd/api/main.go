package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/auth"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/clients"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/handlers"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/metrics"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/vectorizer"
)

const apiPort = "8080"

func main() {
	cfg, err := config.LoadAPIConfig()
	if err != nil {
		logger.Error("invalid api config", "error", err)
		os.Exit(1)
	}

	logger.Info("api server starting",
		"port", apiPort,
		"mongo_uri", env.RedactMongoURI(cfg.MongoURI),
		"redis_addr", cfg.RedisAddr,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	lb, err := clients.NewLoadBalancer(cfg.InferenceTCPAddrs(), cfg.InferenceTimeout)
	if err != nil {
		logger.Error("invalid load balancer config", "error", err)
		os.Exit(1)
	}

	authSvc := auth.New(repo, cfg.JWTSecret, cfg.JWTExpiration)
	h := handlers.New(repo, cache, vec, lb, authSvc, cfg)

	hub := metrics.NewHub(cfg.InferenceHosts, cache, authSvc)
	hubCtx, hubCancel := context.WithCancel(context.Background())
	go hub.Run(hubCtx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", h.Health)
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.Handle("POST /api/v1/predict", authSvc.Middleware(http.HandlerFunc(h.Predict)))
	mux.Handle("POST /api/v1/predict/bulk", authSvc.Middleware(http.HandlerFunc(h.PredictBulk)))
	// mux.Handle("GET /api/v1/metrics", authSvc.Middleware(http.HandlerFunc(h.Metrics)))
	mux.HandleFunc("GET /ws/metrics", hub.HandleWebSocket)

	server := &http.Server{
		Addr:         ":" + apiPort,
		Handler:      handlers.LimitBody(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		hubCancel()
		logger.Info("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("shutdown error", "error", err)
		}
	}()

	logger.Info("http server listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

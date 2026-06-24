package main

import (
	"net"
	"os"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/inference"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
)

func main() {
	cfg, err := config.LoadInferenceConfig()
	if err != nil {
		logger.Error("invalid config", "error", err)
		os.Exit(1)
	}

	logger.Info("inference server starting",
		"port", cfg.Port,
		"mongo_uri", env.RedactMongoURI(cfg.MongoURI),
	)

	model, err := storage.LoadLatestModelFromURI(cfg.MongoURI)
	if err != nil {
		logger.Error("failed to load model", "error", err)
		os.Exit(1)
	}

	logger.Info("model loaded from MongoDB",
		"feature_count", model.FeatureCount,
		"num_classes", model.NumClasses,
		"trained", model.Trained,
	)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("inference server ready", "address", listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("failed to accept connection", "error", err)
			continue
		}
		go inference.HandleConnection(conn, model)
	}
}

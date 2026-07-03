package main

import (
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/inference"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/metrics"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
)

const (
	tcpPort     = "9001"
	metricsPort = "9002"
)

type metricsResponse struct {
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryBytes    uint64  `json:"memory_bytes"`
	RequestsServed int64   `json:"requests_served"`
	CPUName        string  `json:"cpu_name"`
	Cores          int     `json:"cores"`
}

func main() {
	cfg, err := config.LoadInferenceConfig()
	if err != nil {
		logger.Error("invalid config", "error", err)
		os.Exit(1)
	}

	logger.Info("inference server starting",
		"tcp_port", tcpPort,
		"metrics_port", metricsPort,
		"model_path", cfg.ModelPath,
	)

	data, err := os.ReadFile(cfg.ModelPath)
	if err != nil {
		logger.Error("failed to load model", "error", err)
		os.Exit(1)
	}

	var model ml.Model
	if err := json.Unmarshal(data, &model); err != nil {
		logger.Error("failed to parse model", "error", err)
		os.Exit(1)
	}

	logger.Info("model loaded",
		"feature_count", model.FeatureCount,
		"num_classes", model.NumClasses,
		"trained", model.Trained,
	)

	sysInfo, err := metrics.GetSystemInfo()
	if err != nil {
		logger.Warn("failed to get system info, metrics will be limited", "error", err)
	}

	var requestsServed inference.RequestCounter

	// TCP listener
	listener, err := net.Listen("tcp", ":"+tcpPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info("inference server ready", "address", listener.Addr().String())

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Error("failed to accept connection", "error", err)
				continue
			}
			go inference.HandleConnection(conn, &model, &requestsServed)
		}
	}()

	// HTTP metrics server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		proc, err := metrics.GetProcessMetrics()
		if err != nil {
			logger.Warn("failed to get process metrics", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := metricsResponse{
			CPUPercent:     proc.CPUPercent,
			MemoryBytes:    proc.MemoryBytes,
			RequestsServed: requestsServed.Value(),
			CPUName:        sysInfo.CPUBrand,
			Cores:          sysInfo.Cores,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	logger.Info("metrics server starting", "port", metricsPort)
	if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
		logger.Error("metrics server error", "error", err)
		os.Exit(1)
	}
}

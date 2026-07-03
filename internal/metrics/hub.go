package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/auth"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const pollInterval = 2 * time.Second

// Info (static) structs
type infoNode struct {
	Addr       string  `json:"addr"`
	CPUName    string  `json:"cpu_name"`
	Cores      int     `json:"cores"`
	TotalRAMGB float64 `json:"total_ram_gb"`
}

type infoMessage struct {
	Type  string     `json:"type"`
	API   infoNode   `json:"api"`
	Nodes []infoNode `json:"nodes"`
}

// Metrics (dynamic) structs
type nodeMetrics struct {
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryBytes    uint64  `json:"memory_bytes"`
	RequestsServed int64   `json:"requests_served"`
}

type nodePayload struct {
	Addr   string `json:"addr"`
	Status string `json:"status"`
	nodeMetrics
}

type apiMetrics struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryBytes uint64  `json:"memory_bytes"`
}

type clusterMetrics struct {
	TotalPredictions int64   `json:"total_predictions"`
	AvgLatencyMs     float64 `json:"avg_latency_ms"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	TotalNodes       int     `json:"total_nodes"`
	HealthyNodes     int     `json:"healthy_nodes"`
}

type metricsPayload struct {
	Type    string         `json:"type"`
	API     apiMetrics     `json:"api"`
	Nodes   []nodePayload  `json:"nodes"`
	Cluster clusterMetrics `json:"cluster"`
}

// Raw node response
type rawNodeMetrics struct {
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryBytes    uint64  `json:"memory_bytes"`
	RequestsServed int64   `json:"requests_served"`
	CPUName        string  `json:"cpu_name"`
	Cores          int     `json:"cores"`
	TotalRAMGB     float64 `json:"total_ram_gb"`
}

type Hub struct {
	metricsAddrs []string
	hosts        []string
	cache        *storage.Cache
	authSvc      *auth.Auth
	httpClient   *http.Client
	systemInfo   SystemInfo
	clients      map[chan []byte]struct{}
	mu           sync.Mutex
}

func NewHub(hosts []string, cache *storage.Cache, authSvc *auth.Auth) *Hub {
	sysInfo, err := GetSystemInfo()
	if err != nil {
		logger.Warn("failed to get system info for hub", "error", err)
	}

	addrs := make([]string, len(hosts))
	for i, h := range hosts {
		addrs[i] = h + ":9002"
	}

	return &Hub{
		metricsAddrs: addrs,
		hosts:        hosts,
		cache:        cache,
		authSvc:      authSvc,
		httpClient:   &http.Client{Timeout: 2 * time.Second},
		systemInfo:   sysInfo,
		clients:      make(map[chan []byte]struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if data := h.collect(); data != nil {
				h.broadcast(data)
			}
		}
	}
}

func (h *Hub) collect() []byte {
	payload := metricsPayload{
		Type:  "metrics",
		Nodes: make([]nodePayload, 0, len(h.metricsAddrs)),
	}

	// API process metrics
	if proc, err := GetProcessMetrics(); err == nil {
		payload.API.CPUPercent = proc.CPUPercent
		payload.API.MemoryBytes = proc.MemoryBytes
	}

	// Poll inference nodes
	healthy := 0
	for i, addr := range h.metricsAddrs {
		node := nodePayload{
			Addr:   h.hosts[i],
			Status: "unreachable",
		}
		if data, err := h.fetchNodeMetrics(addr); err == nil {
			node.Status = "healthy"
			node.CPUPercent = data.CPUPercent
			node.MemoryBytes = data.MemoryBytes
			node.RequestsServed = data.RequestsServed
			healthy++
		}
		payload.Nodes = append(payload.Nodes, node)
	}

	payload.Cluster.TotalNodes = len(h.hosts)
	payload.Cluster.HealthyNodes = healthy

	// Redis counters
	if h.cache != nil {
		ctx := context.Background()
		if v, err := h.cache.GetCounter(ctx, "predictions_count"); err == nil {
			payload.Cluster.TotalPredictions = v
		}
		if v, err := h.cache.GetFloat64(ctx, "latency_sum"); err == nil {
			if payload.Cluster.TotalPredictions > 0 {
				payload.Cluster.AvgLatencyMs = v / float64(payload.Cluster.TotalPredictions)
			}
		}
		var cacheHits, cacheMisses int64
		if v, err := h.cache.GetCounter(ctx, "cache_hits"); err == nil {
			cacheHits = v
		}
		if v, err := h.cache.GetCounter(ctx, "cache_misses"); err == nil {
			cacheMisses = v
		}
		total := cacheHits + cacheMisses
		if total > 0 {
			payload.Cluster.CacheHitRate = float64(cacheHits) / float64(total)
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.Warn("failed to marshal metrics payload", "error", err)
		return nil
	}
	return data
}

func (h *Hub) fetchNodeMetrics(addr string) (*rawNodeMetrics, error) {
	resp, err := h.httpClient.Get("http://" + addr + "/metrics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var nm rawNodeMetrics
	if err := json.NewDecoder(resp.Body).Decode(&nm); err != nil {
		return nil, err
	}
	return &nm, nil
}

func (h *Hub) collectInfo() []byte {
	payload := infoMessage{
		Type: "info",
		API: infoNode{
			CPUName:    h.systemInfo.CPUBrand,
			Cores:      h.systemInfo.Cores,
			TotalRAMGB: h.systemInfo.TotalRAMGB,
		},
		Nodes: make([]infoNode, 0, len(h.hosts)),
	}

	for i, addr := range h.metricsAddrs {
		node := infoNode{
			Addr: h.hosts[i],
		}

		if data, err := h.fetchNodeMetrics(addr); err == nil {
			node.CPUName = data.CPUName
			node.Cores = data.Cores
			node.TotalRAMGB = data.TotalRAMGB
		}

		payload.Nodes = append(payload.Nodes, node)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.Warn("failed to marshal info payload", "error", err)
		return nil
	}
	return data
}

func (h *Hub) broadcast(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.clients {
		select {
		case ch <- data:
		default:
		}
	}
}

func (h *Hub) addClient(ch chan []byte) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) removeClient(ch chan []byte) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if err := h.authSvc.ValidateToken(token); err != nil {
		protocol.WriteJSON(w, http.StatusUnauthorized, protocol.ErrorResponse{Error: "invalid token"})
		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		logger.Warn("websocket accept failed", "error", err)
		return
	}

	ch := make(chan []byte, 1)
	h.addClient(ch)

	// Send static info on connect
	if info := h.collectInfo(); info != nil {
		select {
		case ch <- info:
		default:
		}
	}

	// Read loop
	go func() {
		defer c.CloseNow()
		defer h.removeClient(ch)
		defer close(ch)

		for {
			_, _, err := c.Read(r.Context())
			if err != nil {
				return
			}
		}
	}()

	// Write loop
	for data := range ch {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		err := wsjson.Write(ctx, c, json.RawMessage(data))
		cancel()
		if err != nil {
			return
		}
	}
}

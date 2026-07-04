package clients

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
)

type LoadBalancer struct {
	addrs   []string
	next    int
	mu      sync.Mutex
	timeout time.Duration
}

func NewLoadBalancer(addrs []string, timeout time.Duration) (*LoadBalancer, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("at least one inference address is required")
	}
	return &LoadBalancer{
		addrs:   addrs,
		timeout: timeout,
	}, nil
}

func (lb *LoadBalancer) Predict(req *protocol.InferenceRequest) (*protocol.InferenceResponse, error) {
	addr := lb.nextAddr()

	var lastErr error
	for i := 0; i < len(lb.addrs); i++ {
		conn, err := net.DialTimeout("tcp", addr, lb.timeout)
		if err != nil {
			logger.Warn("inference node unreachable", "addr", addr, "error", err)
			lastErr = fmt.Errorf("dial %s: %w", addr, err)
			addr = lb.nextAddr()
			continue
		}

		if err := protocol.WriteMessage(conn, req); err != nil {
			conn.Close()
			logger.Warn("inference node write failed", "addr", addr, "error", err)
			lastErr = fmt.Errorf("write to %s: %w", addr, err)
			addr = lb.nextAddr()
			continue
		}

		var resp protocol.InferenceResponse
		if err := protocol.ReadMessage(conn, &resp); err != nil {
			conn.Close()
			logger.Warn("inference node read failed", "addr", addr, "error", err)
			lastErr = fmt.Errorf("read from %s: %w", addr, err)
			addr = lb.nextAddr()
			continue
		}

		conn.Close()
		return &resp, nil
	}

	return nil, fmt.Errorf("all inference nodes failed: %w", lastErr)
}

func (lb *LoadBalancer) PredictScatter(records []protocol.PredictRecord) (*protocol.InferenceResponse, error) {
	n := len(lb.addrs)
	if n <= 1 || len(records) == 0 {
		return lb.Predict(&protocol.InferenceRequest{Records: records})
	}

	chunks := splitRecords(records, n)

	type result struct {
		resp *protocol.InferenceResponse
		err  error
	}
	results := make([]result, n)
	var wg sync.WaitGroup

	for i, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		wg.Add(1)
		go func(idx int, recs []protocol.PredictRecord) {
			defer wg.Done()
			resp, err := lb.Predict(&protocol.InferenceRequest{Records: recs})
			results[idx] = result{resp: resp, err: err}
		}(i, chunk)
	}
	wg.Wait()

	merged := make([]protocol.PredictionResult, 0, len(records))
	for _, r := range results {
		if r.err != nil {
			return nil, fmt.Errorf("scatter-gather: %w", r.err)
		}
		merged = append(merged, r.resp.Predictions...)
	}

	return &protocol.InferenceResponse{Predictions: merged}, nil
}

func splitRecords(records []protocol.PredictRecord, n int) [][]protocol.PredictRecord {
	chunks := make([][]protocol.PredictRecord, n)
	base := len(records) / n
	remainder := len(records) % n
	offset := 0
	for i := range n {
		size := base
		if i < remainder {
			size++
		}
		chunks[i] = records[offset : offset+size]
		offset += size
	}
	return chunks
}

func (lb *LoadBalancer) nextAddr() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	addr := lb.addrs[lb.next]
	lb.next = (lb.next + 1) % len(lb.addrs)
	return addr
}

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

func NewLoadBalancer(addrs []string, timeout time.Duration) *LoadBalancer {
	return &LoadBalancer{
		addrs:   addrs,
		timeout: timeout,
	}
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

func (lb *LoadBalancer) nextAddr() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	addr := lb.addrs[lb.next]
	lb.next = (lb.next + 1) % len(lb.addrs)
	return addr
}

package main

import (
	"context"
	"maps"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Allow(ip string) bool
	RunCleanup(ctx context.Context)
}

type rateLimiter struct {
	cfg         *rateLimiterCfg
	mu          sync.Mutex
	clients     map[string]*client
	rebuildedAt time.Time
}

type rateLimiterCfg struct {
	enable      bool
	rps         float64
	burst       int
	cleanupFreq time.Duration
	rebuildFreq time.Duration
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(cfg rateLimiterCfg) RateLimiter {
	rl := &rateLimiter{
		cfg:         &cfg,
		clients:     make(map[string]*client),
		rebuildedAt: time.Now(),
	}
	return rl
}

func (r *rateLimiter) RunCleanup(ctx context.Context) {
	clientsTicker := time.NewTicker(r.cfg.cleanupFreq)
	rebuildTicker := time.NewTicker(r.cfg.rebuildFreq)
	defer clientsTicker.Stop()
	defer rebuildTicker.Stop()

	for {
		select {
		case <-clientsTicker.C:
			r.cleanupInactive()
		case <-rebuildTicker.C:
			r.rebuildMap()
		case <-ctx.Done():
			return
		}
	}
}

func (r *rateLimiter) rebuildMap() {
	r.mu.Lock()
	defer r.mu.Unlock()

	newMap := make(map[string]*client, len(r.clients))
	maps.Copy(newMap, r.clients)
	r.clients = newMap
}

func (r *rateLimiter) cleanupInactive() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for ip, c := range r.clients {
		if time.Since(c.lastSeen) > r.cfg.cleanupFreq {
			delete(r.clients, ip)
		}
	}
}

func (r *rateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.clients[ip]
	if !ok {
		c = &client{
			limiter:  rate.NewLimiter(r.getRPS(), r.getBurst()),
			lastSeen: time.Now(),
		}
		r.clients[ip] = c
	}

	c.lastSeen = time.Now()

	return c.limiter.Allow()
}

func (r *rateLimiter) getRPS() rate.Limit {
	return rate.Limit(r.cfg.rps)
}

func (r *rateLimiter) getBurst() int {
	return r.cfg.burst
}

type MockLimiter struct{}

func NewMockLimiter() RateLimiter {
	return &MockLimiter{}
}

func (m *MockLimiter) Allow(_ string) bool {
	return true
}

func (m *MockLimiter) IsEnabled() bool {
	return false
}

func (m *MockLimiter) RunCleanup(_ context.Context) {}
